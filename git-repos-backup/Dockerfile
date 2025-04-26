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

VOLUME /data
WORKDIR /data

ENTRYPOINT ["git-repos-backup"]
CMD ["--help"]
