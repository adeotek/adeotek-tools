FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache make git
RUN make build

FROM alpine:3.21

RUN apk add --no-cache git

WORKDIR /app
COPY --from=builder /app/git-repos-backup /usr/local/bin/
COPY --from=builder /app/config.yaml.example /app/config.yaml.example
COPY --from=builder /app/backup_worker.sh /app/backup_worker.sh

RUN chmod +x /app/backup_worker.sh

VOLUME /data

# Environment variables for config
ENV GB_BACKUP_INTERVAL="86400"
ENV GB_VERBOSE="false"

# USER 99:98
# Default to worker script which can handle both config modes
ENTRYPOINT ["/app/backup_worker.sh"]
