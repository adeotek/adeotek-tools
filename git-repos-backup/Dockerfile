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

# For config file mode
ENV GB_CONFIG=""

# For command-line mode
ENV GB_PROVIDER=""
ENV GB_TARGET_DIR=""
ENV GB_SERVER_URL=""
# ENV GB_TOKEN=""
ENV GB_USERNAME=""
# ENV GB_PASSWORD=""
# ENV GB_USE_BASIC_AUTH="false"
ENV GB_SKIP_SSL_VALIDATION="false"
ENV GB_INCLUDE_REPOS=""
ENV GB_EXCLUDE_REPOS=""

# USER 99:98
# Default to worker script which can handle both config modes
ENTRYPOINT ["/app/backup_worker.sh"]
