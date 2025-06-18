package main

import (
	"encoding/base64" // For decoding base64 content
	"flag"            // For parsing command-line arguments
	"fmt"             // For formatted I/O (printing messages)
	"io"              // For I/O primitives (like copying streams)
	"log"             // For logging errors and messages
	"mime"            // For parsing MIME media types and parameters, and WordDecoder
	"mime/multipart"  // For parsing multipart email bodies
	"net/mail"        // For parsing email messages
	"os"              // For operating system interactions (file system)
	"path/filepath"   // For manipulating file paths
	"strings"         // For string manipulation (e.g., checking file extensions)

	// Import character set encodings from golang.org/x/text/encoding

	"golang.org/x/text/encoding/htmlindex" // For looking up charsets by name
	"golang.org/x/text/transform"          // For transforming data streams
)

func main() {
	// Define command-line flags for input and output directories
	inputDir := flag.String("input", "", "Path to the input folder containing .eml files")
	outputDir := flag.String("output", "extracted_pdfs", "Path to the output folder for extracted PDFs")
	flag.Parse() // Parse the command-line arguments

	// Validate that the input directory is provided
	if *inputDir == "" {
		log.Fatal("Error: Input directory is required. Use -input flag.")
	}

	// Check if the input directory exists
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		log.Fatalf("Error: Input directory '%s' does not exist.", *inputDir)
	}

	// Create the output directory if it doesn't already exist.
	// 0755 grants read/write/execute for owner, read/execute for group and others.
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory '%s': %v", *outputDir, err)
	}

	fmt.Printf("Scanning '%s' for .eml files and extracting PDFs to '%s'\n", *inputDir, *outputDir)

	extractedCount := 0 // Counter for extracted PDF files

	// Walk through the input directory recursively
	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return err // Return the error to stop walking
		}
		if info.IsDir() {
			return nil // Skip directories and continue walking
		}

		// Check if the file has a .eml extension (case-insensitive)
		if strings.ToLower(filepath.Ext(path)) == ".eml" {
			log.Printf("Processing EML file: %s\n", path)
			// Call the function to extract PDFs from the current EML file
			count, err := extractPdfsFromEml(path, *outputDir)
			if err != nil {
				log.Printf("Error processing %s: %v\n", path, err)
			}
			extractedCount += count // Accumulate the count of extracted PDFs
		}
		return nil // Continue walking
	})

	if err != nil {
		log.Fatalf("Error walking the directory: %v\n", err)
	}

	fmt.Printf("Finished! Extracted %d PDF(s).\n", extractedCount)
}

// createCharsetReader returns a CharsetReader function suitable for mime.WordDecoder.
// This function maps charsets to their corresponding decoders from golang.org/x/text/encoding.
func createCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	// Look up the encoding by its name (charset string).
	enc, err := htmlindex.Get(charset)
	if err != nil {
		return nil, fmt.Errorf("unhandled charset %q: %w", charset, err)
	}
	return transform.NewReader(input, enc.NewDecoder()), nil
}

// extractPdfsFromEml parses an EML file and extracts PDF attachments.
// It returns the number of PDFs extracted from this file and any error encountered.
func extractPdfsFromEml(emlFilePath, outputDir string) (int, error) {
	file, err := os.Open(emlFilePath)
	if err != nil {
		return 0, fmt.Errorf("could not open EML file %s: %w", emlFilePath, err)
	}
	defer file.Close() // Ensure the file is closed when the function exits

	msg, err := mail.ReadMessage(file)
	if err != nil {
		return 0, fmt.Errorf("could not read EML message from %s: %w", emlFilePath, err)
	}

	// Parse the Content-Type header of the main message
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		// Log a warning if Content-Type is unparseable but don't fail,
		// as it might be a simple text email or a malformed header.
		// Try to process the main body in case it's a direct PDF.
		log.Printf("Warning: Could not parse Content-Type for %s: %v\n", emlFilePath, err)
		return processBody(msg.Body, msg.Header, emlFilePath, outputDir)
	}

	// Check if the message is a multipart message (e.g., multipart/mixed, multipart/alternative)
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"] // Get the boundary string for multipart messages
		if boundary == "" {
			return 0, fmt.Errorf("multipart message without boundary in %s", emlFilePath)
		}

		// Create a new multipart reader from the message body
		mr := multipart.NewReader(msg.Body, boundary)
		extractedInEml := 0 // Counter for PDFs extracted from this specific EML file

		// Iterate through each part of the multipart message
		for {
			p, err := mr.NextPart() // Get the next part
			if err == io.EOF {
				break // No more parts
			}
			if err != nil {
				return extractedInEml, fmt.Errorf("error reading multipart part from %s: %w", emlFilePath, err)
			}

			// Process each part to check for PDF attachments
			count, err := processPart(p, emlFilePath, outputDir)
			if err != nil {
				// Log warnings for individual part errors but continue processing other parts
				log.Printf("Warning: Error processing part in %s: %v\n", emlFilePath, err)
			}
			extractedInEml += count // Add to the count for this EML file
		}
		return extractedInEml, nil
	} else {
		// If it's not a multipart message, try to process the main body itself
		// This handles cases where the entire email content might be a PDF.
		return processBody(msg.Body, msg.Header, emlFilePath, outputDir)
	}
}

// processPart checks a multipart.Part for PDF attachments and saves them.
// It returns 1 if a PDF was extracted, 0 otherwise, and an error if saving failed.
func processPart(p *multipart.Part, emlFilePath, outputDir string) (int, error) {
	// Parse Content-Type header of the part
	contentType, contentTypeParams, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
	if err != nil {
		return 0, fmt.Errorf("could not parse Content-Type for part: %w", err)
	}

	// Parse Content-Disposition header of the part
	disposition, dispParams, err := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
	if err != nil {
		// Log a warning if Content-Disposition is unparseable but continue processing.
		// We will try to derive filename from Content-Type if disposition is problematic.
		log.Printf("Warning: Could not parse Content-Disposition for part in %s (header: '%s'): %v\n",
			emlFilePath, p.Header.Get("Content-Disposition"), err)
		disposition = "" // Clear disposition if parsing failed, so logic below can use it as "not attachment"
	}

	// Check if it's a PDF. We consider it an attachment if Content-Disposition is "attachment"
	// OR if Content-Type is "application/pdf" AND it has a 'name' parameter (common for inline attachments).
	isPdfAttachment := contentType == "application/pdf" &&
		(disposition == "attachment" || (disposition == "" && contentTypeParams["name"] != ""))

	if isPdfAttachment {
		filename := dispParams["filename"] // Try to get filename from Content-Disposition first

		if filename == "" {
			// Fallback: Try to get filename from 'name' parameter in Content-Type
			filename = contentTypeParams["name"]
		}

		// Decode the filename if it uses MIME "Encoded-Word Syntax" (e.g., =?UTF-8?B?...)
		// Trim whitespace before decoding to handle potential hidden characters.
		filenameToDecode := strings.TrimSpace(filename)

		// Create a new WordDecoder instance and set its CharsetReader.
		decoder := &mime.WordDecoder{
			CharsetReader: createCharsetReader,
		}

		decodedFilename, err := decoder.DecodeHeader(filenameToDecode)
		if err == nil { // If decoding is successful, use the decoded filename
			filename = decodedFilename
		} else {
			// Log the specific error from DecodeHeader for better debugging
			log.Printf("Warning: mime.WordDecoder failed to decode filename '%s' from part in %s: %v. Using original filename.\n",
				filenameToDecode, emlFilePath, err)
			// The 'filename' variable will retain its original (encoded) value if decoding fails.
		}

		if filename == "" {
			// If no filename is provided or after decoding it's empty, generate a fallback name
			filename = fmt.Sprintf("unnamed_pdf_%s.pdf", strings.ReplaceAll(filepath.Base(emlFilePath), ".", "_"))
			log.Printf("Warning: PDF attachment in %s has no filename, using '%s'\n", emlFilePath, filename)
		}

		// Determine the appropriate reader for the part's content
		var partReader io.Reader = p
		transferEncoding := p.Header.Get("Content-Transfer-Encoding")
		if strings.ToLower(transferEncoding) == "base64" {
			// If content is base64 encoded, decode it on the fly
			partReader = base64.NewDecoder(base64.StdEncoding, p)
		}

		// Construct the full path for the output PDF file
		outputFilePath := filepath.Join(outputDir, filename)
		// Ensure the filename is unique to avoid overwriting existing files
		uniqueFilePath := getUniqueFilename(outputFilePath)

		// Create the output file
		outFile, err := os.Create(uniqueFilePath)
		if err != nil {
			return 0, fmt.Errorf("could not create output file %s: %w", uniqueFilePath, err)
		}
		defer outFile.Close() // Close the output file when done

		// Copy the content from the part reader to the output file
		bytesWritten, err := io.Copy(outFile, partReader)
		if err != nil {
			return 0, fmt.Errorf("could not write PDF content to %s: %w", uniqueFilePath, err)
		}

		log.Printf("Extracted PDF: %s (%d bytes)\n", uniqueFilePath, bytesWritten)
		return 1, nil // Return 1 indicating one PDF was extracted
	}
	return 0, nil // Not a PDF attachment, return 0
}

// processBody attempts to process a non-multipart message body for a PDF.
// This is used if the entire EML's body is a PDF, not as a multipart attachment.
func processBody(body io.Reader, headers mail.Header, emlFilePath, outputDir string) (int, error) {
	// Parse Content-Type header of the main message body
	contentType, contentTypeParams, err := mime.ParseMediaType(headers.Get("Content-Type"))
	if err != nil {
		return 0, nil // No content type or unparseable, assume not a PDF
	}

	// Parse Content-Disposition header of the main message body (might be empty)
	disposition, dispParams, err := mime.ParseMediaType(headers.Get("Content-Disposition"))
	if err != nil {
		log.Printf("Warning: Could not parse Content-Disposition for main body in %s (header: '%s'): %v\n",
			emlFilePath, headers.Get("Content-Disposition"), err)
		disposition = "" // Clear disposition if parsing failed
	}

	// Check if the main body is a PDF and (is an attachment or has no disposition specified
	// but has a 'name' parameter in Content-Type).
	isPdfAttachment := contentType == "application/pdf" &&
		(disposition == "attachment" || (disposition == "" && contentTypeParams["name"] != ""))

	if isPdfAttachment {
		filename := dispParams["filename"] // Try from Content-Disposition first

		if filename == "" {
			// Fallback: Try to get filename from 'name' parameter in Content-Type
			filename = contentTypeParams["name"]
		}

		// Decode the filename if it uses MIME "Encoded-Word Syntax"
		// Trim whitespace before decoding.
		filenameToDecode := strings.TrimSpace(filename)

		// Create a new WordDecoder instance and set its CharsetReader.
		decoder := &mime.WordDecoder{
			CharsetReader: createCharsetReader,
		}

		decodedFilename, err := decoder.DecodeHeader(filenameToDecode)
		if err == nil {
			filename = decodedFilename
		} else {
			// Log the specific error from DecodeHeader for better debugging
			log.Printf("Warning: mime.WordDecoder failed to decode filename '%s' from main body in %s: %v. Using original filename.\n",
				filenameToDecode, emlFilePath, err)
			// The 'filename' variable will retain its original (encoded) value if decoding fails.
		}

		if filename == "" {
			// Generate a fallback name for the PDF from the main body
			filename = fmt.Sprintf("unnamed_body_pdf_%s.pdf", strings.ReplaceAll(filepath.Base(emlFilePath), ".", "_"))
			log.Printf("Warning: Main body PDF in %s has no filename, using '%s'\n", emlFilePath, filename)
		}

		// Determine the appropriate reader for the body's content
		var bodyReader io.Reader = body
		transferEncoding := headers.Get("Content-Transfer-Encoding")
		if strings.ToLower(transferEncoding) == "base64" {
			bodyReader = base64.NewDecoder(base64.StdEncoding, body)
		}

		// Construct the full path for the output PDF file
		outputFilePath := filepath.Join(outputDir, filename)
		// Ensure the filename is unique
		uniqueFilePath := getUniqueFilename(outputFilePath)

		// Create the output file
		outFile, err := os.Create(uniqueFilePath)
		if err != nil {
			return 0, fmt.Errorf("could not create output file %s: %w", uniqueFilePath, err)
		}
		defer outFile.Close()

		// Copy the content from the body reader to the output file
		bytesWritten, err := io.Copy(outFile, bodyReader)
		if err != nil {
			return 0, fmt.Errorf("could not write PDF content to %s: %w", uniqueFilePath, err)
		}

		log.Printf("Extracted PDF from main body: %s (%d bytes)\n", uniqueFilePath, bytesWritten)
		return 1, nil // Return 1 indicating one PDF was extracted
	}
	return 0, nil // Not a PDF in the main body, return 0
}

// getUniqueFilename appends a counter to the filename if a file with the same name already exists.
func getUniqueFilename(filePath string) string {
	ext := filepath.Ext(filePath)             // Get file extension (e.g., ".pdf")
	base := filePath[:len(filePath)-len(ext)] // Get base name without extension
	counter := 1                              // Start counter for uniqueness

	for {
		_, err := os.Stat(filePath) // Check if the file exists
		if os.IsNotExist(err) {
			return filePath // File does not exist, so the current path is unique
		}
		// File exists, construct a new path with a counter (e.g., "file_1.pdf")
		filePath = fmt.Sprintf("%s_%d%s", base, counter, ext)
		counter++ // Increment counter for the next attempt
	}
}
