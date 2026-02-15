FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o sql-toolbox ./cmd/sql-toolbox


FROM alpine:latest

# Install PostgreSQL client for pg_dump support
RUN apk update && apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=builder /build/sql-toolbox /usr/local/bin/
COPY --from=builder /build/backup_worker.sh /app/
COPY --from=builder /build/config.yaml.example /app/config.yaml.example

RUN chmod +x /app/backup_worker.sh

# Data volume
VOLUME /data

# Environment variables for config
ENV DBB_CONFIG=""
ENV DBB_BACKUP_INTERVAL="86400"
ENV DBB_VERBOSE="false"

# USER 99:98
# Default to worker script which can handle both config modes
ENTRYPOINT ["/app/backup_worker.sh"]
