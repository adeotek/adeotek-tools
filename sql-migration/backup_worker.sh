#!/bin/sh

DEFAULT_CONFIG_FILE="/data/config.yaml"
EXAMPLE_CONFIG_FILE="/app/config.yaml.example"
DBB_SLEEP_TIME=${DBB_BACKUP_INTERVAL:-86400}  # Default to 24 hours (86400 seconds)

echo "Starting adeotek-tools/sql-migration--backup-only"
echo "Author: adeotek"
echo "Target: Unraid"
echo ""

if [[ -n "$DBB_CONFIG" ]]; then
  if [[ ! -f "$DBB_CONFIG" ]]; then
    echo "Config file [$DBB_CONFIG] not found, exiting..."
    exit 1
  fi

  CONFIG_FILE="$DBB_CONFIG"
else
  if [[ ! -f "$DEFAULT_CONFIG_FILE" ]]; then
    if [[ ! -f "$EXAMPLE_CONFIG_FILE" ]]; then
      echo "Config file not found and no example config available, exiting..."
      exit 1
    fi

    echo "Config file not found, creating from example..."
    mkdir -p /data
    cp "$EXAMPLE_CONFIG_FILE" "$DEFAULT_CONFIG_FILE"
  fi

  CONFIG_FILE="$DEFAULT_CONFIG_FILE"
fi

BACKUP_CMD="sql-migration backup-only --config $CONFIG_FILE"
# Add verbose flag if set
if [ "$DBB_VERBOSE" = "true" ]; then
  BACKUP_CMD="$BACKUP_CMD --verbose"
fi

echo "Starting backup only worker with interval: $DBB_SLEEP_TIME seconds"
echo "Using command: $BACKUP_CMD"

# Run backup in infinite loop
while true; do
  echo "Starting backup at $(date) > $BACKUP_CMD"
  $BACKUP_CMD
  echo "Backup completed at $(date), sleeping for $DBB_SLEEP_TIME seconds"
  sleep $DBB_SLEEP_TIME
done
