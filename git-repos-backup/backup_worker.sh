#!/bin/sh

CONFIG_FILE="${GB_CONFIG:-/app/config.yaml}"
GB_SLEEP_TIME=${GB_BACKUP_INTERVAL:-86400}  # Default to 24 hours (86400 seconds)

echo "Starting adeotek-tools/git-repos-backup"
echo "Author: adeotek"
echo "Target: Unraid"
echo ""

# Store each argument separately so that values containing spaces or shell
# metacharacters are passed through literally instead of being word-split.
set -- git-repos-backup

# Choose between config file and command-line arguments
if [ -n "$GB_PROVIDER" ] && [ -n "$GB_TARGET_DIR" ]; then
  echo "Using command-line arguments from environment variables"
  set -- "$@" -provider "$GB_PROVIDER" -target-dir "$GB_TARGET_DIR"

  if [ -n "$GB_SERVER_URL" ]; then
    set -- "$@" -server-url "$GB_SERVER_URL"
  fi

  if [ -n "$GB_TOKEN" ]; then
    set -- "$@" -token "$GB_TOKEN"
  fi

  if [ -n "$GB_USERNAME" ]; then
    set -- "$@" -username "$GB_USERNAME"
  fi

  if [ -n "$GB_PASSWORD" ]; then
    set -- "$@" -password "$GB_PASSWORD"
  fi

  if [ "$GB_USE_BASIC_AUTH" = "true" ]; then
    set -- "$@" -use-basic-auth
  fi

  if [ "$GB_SKIP_SSL_VALIDATION" = "true" ]; then
    set -- "$@" -skip-ssl
  fi

  if [ -n "$GB_INCLUDE_REPOS" ]; then
    set -- "$@" -include "$GB_INCLUDE_REPOS"
  fi

  if [ -n "$GB_EXCLUDE_REPOS" ]; then
    set -- "$@" -exclude "$GB_EXCLUDE_REPOS"
  fi
else
  if [ ! -f "$CONFIG_FILE" ]; then
    echo "Config file not found: $CONFIG_FILE, exiting..."
    exit 1
  fi
  set -- "$@" -config "$CONFIG_FILE"
fi

# Add verbose flag if set
if [ "$GB_VERBOSE" = "true" ]; then
  set -- "$@" -verbose
fi

echo "Starting backup worker with interval: $GB_SLEEP_TIME seconds"
echo "Using command: $*"

# Run backup in infinite loop
while true; do
  echo "Starting backup at $(date)"
  "$@"
  echo "Backup completed at $(date), sleeping for $GB_SLEEP_TIME seconds"
  sleep "$GB_SLEEP_TIME"
done
