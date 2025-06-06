#!/bin/sh

CONFIG_FILE="/app/config.yaml"
CONFIG_EXAMPLE="/app/config.yaml.example"
GB_SLEEP_TIME=${GB_BACKUP_INTERVAL:-86400}  # Default to 24 hours (86400 seconds)

echo "Starting adeotek-tools/git-repos-backup"
echo "Author: adeotek"
echo "Target: Unraid"
echo ""

# Choose between config file and command-line arguments
USE_CONFIG_FILE=1

# If environment variables are set for direct usage, use command-line args instead
if [ -n "$GB_PROVIDER" ] && [ -n "$GB_TARGET_DIR" ]; then
  USE_CONFIG_FILE=0
  echo "Using command-line arguments from environment variables"
fi

if [ "$USE_CONFIG_FILE" -eq 1 ]; then
  # Construct the command with config file
  if [ -f "$GB_CONFIG" ]; then
    BACKUP_CMD="git-repos-backup -config $GB_CONFIG"
  else
    # Check if config file exists, if not create from example
    if [ ! -f "$CONFIG_FILE" ]; then
      echo "Config file not found, exiting..."
      exit 1
    fi
    BACKUP_CMD="git-repos-backup -config $CONFIG_FILE"
  fi
else
  # Construct the command with environment variables
  BACKUP_CMD="git-repos-backup -provider $GB_PROVIDER -target-dir $GB_TARGET_DIR"

  # Add optional parameters if they are set
  if [ -n "$GB_SERVER_URL" ]; then
    BACKUP_CMD="$BACKUP_CMD -server-url $GB_SERVER_URL"
  fi

  if [ -n "$GB_TOKEN" ]; then
    BACKUP_CMD="$BACKUP_CMD -token $GB_TOKEN"
  fi

  if [ -n "$GB_USERNAME" ]; then
    BACKUP_CMD="$BACKUP_CMD -username $GB_USERNAME"
  fi

  if [ -n "$GB_PASSWORD" ]; then
    BACKUP_CMD="$BACKUP_CMD -password $GB_PASSWORD"
  fi

  if [ "$GB_USE_BASIC_AUTH" = "true" ]; then
    BACKUP_CMD="$BACKUP_CMD -use-basic-auth"
  fi

  if [ "$GB_SKIP_SSL_VALIDATION" = "true" ]; then
    BACKUP_CMD="$BACKUP_CMD -skip-ssl"
  fi

  if [ -n "$GB_INCLUDE_REPOS" ]; then
    BACKUP_CMD="$BACKUP_CMD -include $GB_INCLUDE_REPOS"
  fi

  if [ -n "$GB_EXCLUDE_REPOS" ]; then
    BACKUP_CMD="$BACKUP_CMD -exclude $GB_EXCLUDE_REPOS"
  fi
fi

# Add verbose flag if set
if [ "$GB_VERBOSE" = "true" ]; then
  BACKUP_CMD="$BACKUP_CMD -verbose"
fi

echo "Starting backup worker with interval: $GB_SLEEP_TIME seconds"
echo "Using command: $BACKUP_CMD"

# Run backup in infinite loop
while true; do
  echo "Starting backup at $(date) > $BACKUP_CMD"
  $BACKUP_CMD
  echo "Backup completed at $(date), sleeping for $GB_SLEEP_TIME seconds"
  sleep $GB_SLEEP_TIME
done
