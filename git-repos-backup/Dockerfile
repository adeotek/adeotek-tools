FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Print directory structure for debugging
RUN ls -la && ls -la cmd/

# Install build dependencies
RUN apk add --no-cache git

# Build the application with more verbose output
RUN CGO_ENABLED=0 GOOS=linux go build -v -o git-repos-backup ./cmd/git-repos-backup

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy only necessary files from builder
COPY --from=builder /build/git-repos-backup /usr/local/bin/
COPY --from=builder /build/backup_worker.sh /app/
COPY --from=builder /build/config.yaml.example /app/

# Set permissions
RUN chmod +x /app/backup_worker.sh

VOLUME /data

# Environment variables for config
ENV GB_BACKUP_INTERVAL="86400"
ENV GB_VERBOSE="false"

# Default to worker script which can handle both config modes
ENTRYPOINT ["/app/backup_worker.sh"]
