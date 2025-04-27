FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go build -o git-repos-backup ./cmd/git-repos-backup


FROM alpine:latest

RUN apk add --no-cache git

WORKDIR /app

COPY --from=builder /build/git-repos-backup /usr/local/bin/
COPY --from=builder /build/backup_worker.sh /app/
COPY --from=builder /build/config.yaml.example /app/

RUN chmod +x /app/backup_worker.sh

# Data volume
VOLUME /data

# Environment variables for config
ENV GB_BACKUP_INTERVAL="86400"
ENV GB_VERBOSE="false"

ENTRYPOINT ["/app/backup_worker.sh"]
