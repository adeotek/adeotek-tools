#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="claude-code-dashboard"
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
ENV_FILE="$SYSTEMD_USER_DIR/$SERVICE_NAME.env"
UNIT_FILE="$SYSTEMD_USER_DIR/$SERVICE_NAME.service"

# Stop and disable
if systemctl --user is-active --quiet "$SERVICE_NAME" 2>/dev/null || \
   systemctl --user is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
  echo "Stopping and disabling $SERVICE_NAME..."
  systemctl --user disable --now "$SERVICE_NAME" || true
else
  echo "Service $SERVICE_NAME is not active or enabled — skipping stop."
fi

# Remove unit file
if [[ -f "$UNIT_FILE" ]]; then
  rm "$UNIT_FILE"
  echo "Removed: $UNIT_FILE"
else
  echo "Unit file not found — skipping: $UNIT_FILE"
fi

# Prompt before removing env file (may have been customised)
if [[ -f "$ENV_FILE" ]]; then
  read -rp "Remove env file $ENV_FILE? (y/N) " confirm
  if [[ "${confirm,,}" == "y" ]]; then
    rm "$ENV_FILE"
    echo "Removed: $ENV_FILE"
  else
    echo "Kept: $ENV_FILE"
  fi
fi

systemctl --user daemon-reload
echo ""
echo "Done. Linger was NOT disabled (may be needed by other user services)."
echo "To disable linger manually: loginctl disable-linger \$USER"
