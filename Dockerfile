# Build stage
FROM golang:1.21-alpine AS builder

# Install git for private repos and ca-certificates for SSL
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o pdf-from-eml main.go

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/pdf-from-eml /pdf-from-eml

# Create directories for input/output
VOLUME ["/input", "/output"]

# Set default working directory
WORKDIR /

# Default command
ENTRYPOINT ["/pdf-from-eml"]
CMD ["-input", "/input", "-output", "/output"]

# Metadata
LABEL maintainer="fmurodov" \
    description="Extract PDF attachments from EML files" \
    version="1.0.0"
