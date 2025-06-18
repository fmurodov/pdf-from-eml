# PDF from EML Extractor

[![CI/CD Pipeline](https://github.com/fmurodov/pdf-from-eml/actions/workflows/ci.yml/badge.svg)](https://github.com/fmurodov/pdf-from-eml/actions/workflows/ci.yml)
[![Release](https://github.com/fmurodov/pdf-from-eml/actions/workflows/release.yml/badge.svg)](https://github.com/fmurodov/pdf-from-eml/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fmurodov/pdf-from-eml)](https://goreportcard.com/report/github.com/fmurodov/pdf-from-eml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/docker/pulls/fmurodov/pdf-from-eml)](https://hub.docker.com/r/fmurodov/pdf-from-eml)

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

### Option 1: Download Pre-built Binaries (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/fmurodov/pdf-from-eml/releases):

#### Linux (AMD64)
```bash
curl -L -o pdf-from-eml.tar.gz https://github.com/fmurodov/pdf-from-eml/releases/latest/download/pdf-from-eml_linux_amd64.tar.gz
tar -xzf pdf-from-eml.tar.gz
chmod +x pdf-from-eml
sudo mv pdf-from-eml /usr/local/bin/
```

#### macOS (AMD64)
```bash
curl -L -o pdf-from-eml.tar.gz https://github.com/fmurodov/pdf-from-eml/releases/latest/download/pdf-from-eml_darwin_amd64.tar.gz
tar -xzf pdf-from-eml.tar.gz
chmod +x pdf-from-eml
sudo mv pdf-from-eml /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L -o pdf-from-eml.tar.gz https://github.com/fmurodov/pdf-from-eml/releases/latest/download/pdf-from-eml_darwin_arm64.tar.gz
tar -xzf pdf-from-eml.tar.gz
chmod +x pdf-from-eml
sudo mv pdf-from-eml /usr/local/bin/
```

#### Windows (AMD64)
Download the `.zip` file from the releases page and extract it to your desired location.

### Option 2: Install with Go

```bash
go install github.com/fmurodov/pdf-from-eml@latest
```

### Option 3: Docker

```bash
# Run with Docker
docker run --rm -v /path/to/eml/files:/input -v /path/to/output:/output fmurodov/pdf-from-eml:latest

# Or pull the image first
docker pull fmurodov/pdf-from-eml:latest
```

### Option 4: Build from Source

```bash
git clone https://github.com/fmurodov/pdf-from-eml.git
cd pdf-from-eml
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
- `-version`: Show version information

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

Check version:
```bash
./pdf-from-eml -version
```

### Docker Usage

```bash
# Basic usage
docker run --rm \
  -v /path/to/your/eml/files:/input \
  -v /path/to/output/directory:/output \
  fmurodov/pdf-from-eml:latest

# With custom options (not applicable as Docker uses default CMD)
docker run --rm \
  -v /path/to/your/eml/files:/input \
  -v /path/to/output/directory:/output \
  fmurodov/pdf-from-eml:latest \
  -input /input -output /output
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

## Supported Platforms

This tool is built and tested on multiple platforms and architectures:

| OS      | AMD64 | ARM64 | ARM |
|---------|-------|-------|-----|
| Linux   | ✅    | ✅    | ✅  |
| macOS   | ✅    | ✅    | ❌  |
| Windows | ✅    | ✅    | ❌  |

## CI/CD and Quality Assurance

This project uses GitHub Actions for:
- **Continuous Integration**: Automated testing on every commit
- **Cross-platform Building**: Automated builds for all supported platforms
- **Security Scanning**: Automated security vulnerability checks
- **Code Quality**: Static analysis with `go vet` and `staticcheck`
- **Docker Images**: Multi-architecture Docker images
- **Automated Releases**: Tagged releases with pre-built binaries

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