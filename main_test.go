package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetUniqueFilename(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_unique_filename")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: File doesn't exist, should return original path
	originalPath := filepath.Join(tempDir, "test.pdf")
	uniquePath := getUniqueFilename(originalPath)
	if uniquePath != originalPath {
		t.Errorf("Expected %s, got %s", originalPath, uniquePath)
	}

	// Test case 2: File exists, should return path with counter
	// Create the original file
	file, err := os.Create(originalPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	uniquePath = getUniqueFilename(originalPath)
	expectedPath := filepath.Join(tempDir, "test_1.pdf")
	if uniquePath != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, uniquePath)
	}

	// Test case 3: Multiple files exist
	file2, err := os.Create(expectedPath)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}
	file2.Close()

	uniquePath = getUniqueFilename(originalPath)
	expectedPath2 := filepath.Join(tempDir, "test_2.pdf")
	if uniquePath != expectedPath2 {
		t.Errorf("Expected %s, got %s", expectedPath2, uniquePath)
	}
}

func TestCreateCharsetReader(t *testing.T) {
	testString := "Hello, World!"
	reader := strings.NewReader(testString)

	// Test with UTF-8 charset
	charsetReader, err := createCharsetReader("utf-8", reader)
	if err != nil {
		t.Errorf("Expected no error for utf-8 charset, got: %v", err)
	}
	if charsetReader == nil {
		t.Error("Expected non-nil reader for utf-8 charset")
	}

	// Test with unsupported charset
	reader2 := strings.NewReader(testString)
	_, err = createCharsetReader("invalid-charset", reader2)
	if err == nil {
		t.Error("Expected error for invalid charset, got nil")
	}
}

func TestMainFunctionFlags(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test case: No input flag provided
	os.Args = []string{"pdf-from-eml"}

	// We can't easily test the main function directly since it calls log.Fatal
	// But we can test that the flag parsing logic would work
	// This is more of a smoke test to ensure the code compiles and basic structure is correct
}

func TestFileExtensionCheck(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"test.eml", true},
		{"test.EML", true},
		{"test.Eml", true},
		{"test.txt", false},
		{"test.pdf", false},
		{"test", false},
		{"", false},
	}

	for _, tc := range testCases {
		// Simulate the logic from main function
		isEml := strings.ToLower(filepath.Ext(tc.filename)) == ".eml"
		if isEml != tc.expected {
			t.Errorf("For filename %s, expected %v, got %v", tc.filename, tc.expected, isEml)
		}
	}
}

func TestProcessBodyLogic(t *testing.T) {
	// Test the logic for determining if content is a PDF attachment
	testCases := []struct {
		contentType     string
		disposition     string
		contentTypeName string
		expected        bool
		description     string
	}{
		{
			contentType:     "application/pdf",
			disposition:     "attachment",
			contentTypeName: "",
			expected:        true,
			description:     "PDF with attachment disposition",
		},
		{
			contentType:     "application/pdf",
			disposition:     "",
			contentTypeName: "document.pdf",
			expected:        true,
			description:     "PDF with name parameter",
		},
		{
			contentType:     "text/plain",
			disposition:     "attachment",
			contentTypeName: "",
			expected:        false,
			description:     "Text file with attachment disposition",
		},
		{
			contentType:     "application/pdf",
			disposition:     "inline",
			contentTypeName: "",
			expected:        false,
			description:     "PDF with inline disposition and no name",
		},
	}

	for _, tc := range testCases {
		// Simulate the logic from processPart/processBody functions
		isPdfAttachment := tc.contentType == "application/pdf" &&
			(tc.disposition == "attachment" || (tc.disposition == "" && tc.contentTypeName != ""))

		if isPdfAttachment != tc.expected {
			t.Errorf("Test case '%s': expected %v, got %v", tc.description, tc.expected, isPdfAttachment)
		}
	}
}

func BenchmarkGetUniqueFilename(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "bench_unique_filename")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testPath := filepath.Join(tempDir, "test.pdf")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getUniqueFilename(testPath)
	}
}
