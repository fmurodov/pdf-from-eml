# PDF from EML Extractor

A Go command-line tool that extracts PDF attachments from EML (email message) files. This tool recursively scans directories for `.eml` files and extracts any PDF attachments found within them.

## Features

- **Recursive Directory Scanning**: Automatically finds all `.eml` files in a directory and its subdirectories
- **PDF Extraction**: Extracts PDF attachments from email messages
- **Multiple Email Formats**: Supports both multipart and single-part email messages
- **Character Encoding Support**: Handles various character encodings using `golang.org/x/text/encoding`
- **MIME Decoding**: Properly decodes MIME-encoded filenames (e.g., UTF-8, Base64)
- **Base64 Decoding**: Automatically decodes base64-encoded PDF attachments
- **Filename Collision Handling**: Automatically generates unique filenames to prevent overwriting
- **Detailed Logging**: Provides informative output about the extraction process

## Requirements

- Go 1.24.4 or later
- Dependencies (automatically managed via Go modules):
  - `golang.org/x/text v0.26.0`

## Installation

### Option 1: Clone from GitHub

```bash
git clone https://github.com/fmurodov/pdf-from-eml.git
cd pdf-from-eml
go build -o pdf-from-eml main.go
```

### Option 2: Install directly with Go

```bash
go install github.com/fmurodov/pdf-from-eml@latest
```

### Option 3: Download and build manually

1. Download this repository
2. Navigate to the project directory
3. Build the application:

```bash
go build -o pdf-from-eml main.go
```

## Usage

### Basic Usage

```bash
./pdf-from-eml -input /path/to/eml/files
```

### Custom Output Directory

```bash
./pdf-from-eml -input /path/to/eml/files -output /path/to/output/directory
```

### Command-Line Options

- `-input`: **Required**. Path to the input folder containing `.eml` files
- `-output`: Optional. Path to the output folder for extracted PDFs (default: `extracted_pdfs`)

### Examples

Extract PDFs from a single directory:
```bash
./pdf-from-eml -input ./emails
```

Extract PDFs with custom output directory:
```bash
./pdf-from-eml -input ./emails -output ./my_pdfs
```

Extract PDFs from a nested directory structure:
```bash
./pdf-from-eml -input /home/user/Documents/archived_emails -output /home/user/Documents/extracted_pdfs
```

## How It Works

1. **Directory Scanning**: The tool recursively walks through the specified input directory looking for files with `.eml` extensions
2. **Email Parsing**: Each EML file is parsed using Go's `net/mail` package
3. **Content Analysis**: The tool analyzes the email structure:
   - For multipart emails: Examines each part for PDF attachments
   - For single-part emails: Checks if the main body contains a PDF
4. **PDF Detection**: Identifies PDFs by checking:
   - `Content-Type: application/pdf`
   - `Content-Disposition: attachment` headers
   - Inline PDFs with filename parameters
5. **Extraction**: Saves PDF content to the output directory with proper filename handling

## File Naming

- Original filenames from email attachments are preserved when possible
- MIME-encoded filenames are automatically decoded
- If no filename is available, a generated name is used (e.g., `unnamed_pdf_email1_eml.pdf`)
- Duplicate filenames are handled by appending a counter (e.g., `document_1.pdf`, `document_2.pdf`)

## Error Handling

The tool includes robust error handling:
- Invalid or corrupted EML files are logged and skipped
- Parsing errors for individual email parts don't stop the entire process
- File system errors are properly reported
- Character encoding issues are handled gracefully with fallbacks

## Output

The tool provides detailed console output including:
- Progress information for each processed EML file
- Success messages for extracted PDFs with file sizes
- Warning messages for any issues encountered
- Final summary of total PDFs extracted

## Supported Email Formats

- Standard RFC 5322 email messages
- Multipart MIME messages (`multipart/mixed`, `multipart/alternative`, etc.)
- Base64 encoded attachments
- Various character encodings (UTF-8, ISO-8859-1, etc.)
- Inline and attached PDFs

## AI Generated Code Notice

⚠️ **Important**: This project's code and documentation were generated using AI assistance. While the code has been designed to handle various edge cases and follows Go best practices, please review and test thoroughly before using in production environments.

## License

This project is provided as-is for educational and utility purposes. Please ensure you have the right to extract and process the email files you're working with.

## Contributing

Feel free to submit issues, feature requests, or improvements to enhance the tool's functionality.

## Troubleshooting

### Common Issues

1. **"Input directory is required"**: Make sure to specify the `-input` flag with a valid directory path
2. **"Input directory does not exist"**: Verify the path to your EML files is correct
3. **No PDFs extracted**: Check that your EML files actually contain PDF attachments
4. **Permission errors**: Ensure you have read access to the input directory and write access to the output directory

### Debug Tips

- Check the console output for detailed processing information
- Verify that your EML files are valid email message files
- Ensure PDF attachments in the emails have proper MIME types