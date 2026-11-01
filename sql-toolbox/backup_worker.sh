#!/bin/sh

DEFAULT_CONFIG_FILE="/data/config.yaml"
EXAMPLE_CONFIG_FILE="/app/config.yaml.example"
DBB_SLEEP_TIME=${DBB_BACKUP_INTERVAL:-86400} # Default to 24 hours (86400 seconds)

case "$DBB_SLEEP_TIME" in
  ''|*[!0-9]*)
    echo "Invalid DBB_BACKUP_INTERVAL '$DBB_SLEEP_TIME', using 86400 seconds"
    DBB_SLEEP_TIME=86400
    ;;
esac

echo "Starting adeotek-tools/sql-toolbox--backup"
echo "Author: adeotek"
echo "Target: Unraid"
echo ""

if [ -n "$DBB_CONFIG" ]; then
  if [ ! -f "$DBB_CONFIG" ]; then
    echo "Config file [$DBB_CONFIG] not found, exiting..."
    exit 1
  fi

  CONFIG_FILE="$DBB_CONFIG"
else
  if [ ! -f "$DEFAULT_CONFIG_FILE" ]; then
    if [ ! -f "$EXAMPLE_CONFIG_FILE" ]; then
      echo "Config file not found and no example config available, exiting..."
      exit 1
    fi

    echo "Config file not found, creating from example..."
    mkdir -p /data
    cp "$EXAMPLE_CONFIG_FILE" "$DEFAULT_CONFIG_FILE"
  fi

  CONFIG_FILE="$DEFAULT_CONFIG_FILE"
fi

echo "Starting backup only worker with interval: $DBB_SLEEP_TIME seconds"
echo "Using config: $CONFIG_FILE"

# Run backup in an infinite loop, logging the outcome of every run
while true; do
  echo "Starting backup at $(date)"
  if [ "$DBB_VERBOSE" = "true" ]; then
    sql-toolbox backup --config "$CONFIG_FILE" --verbose
  else
    sql-toolbox backup --config "$CONFIG_FILE"
  fi
  status=$?
  if [ "$status" -eq 0 ]; then
    echo "Backup completed successfully at $(date), sleeping for $DBB_SLEEP_TIME seconds"
  else
    echo "Backup failed with exit code $status at $(date), retrying in $DBB_SLEEP_TIME seconds"
  fi
  sleep "$DBB_SLEEP_TIME"
done
