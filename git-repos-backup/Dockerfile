FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Install build dependencies
RUN apk add --no-cache git

# Build the application directly
RUN CGO_ENABLED=0 GOOS=linux go build -o git-repos-backup ./cmd/git-repos-backup

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy only necessary files from builder
COPY --from=builder /app/git-repos-backup /usr/local/bin/
COPY --from=builder /app/backup_worker.sh /app/
COPY --from=builder /app/config.yaml.example /app/

# Set permissions
RUN chmod +x /app/backup_worker.sh

VOLUME /data

# Environment variables for config
ENV GB_BACKUP_INTERVAL="86400"
ENV GB_VERBOSE="false"

# Default to worker script which can handle both config modes
ENTRYPOINT ["/app/backup_worker.sh"]
