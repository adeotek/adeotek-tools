#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
SERVICE_NAME="claude-code-dashboard"
ENV_FILE="$SYSTEMD_USER_DIR/$SERVICE_NAME.env"
UNIT_FILE="$SYSTEMD_USER_DIR/$SERVICE_NAME.service"
TEMPLATE="$SCRIPT_DIR/$SERVICE_NAME.service.template"
DOT_ENV="$INSTALL_DIR/backend/.env"

# --- Preflight ---
check_command() {
  if ! command -v "$1" &>/dev/null; then
    echo "Error: '$1' not found on PATH. Please install it and try again." >&2
    exit 1
  fi
}
check_command node
check_command npm
check_command claude

NODE_BIN="$(command -v node)"

# --- Build ---
SKIP_BUILD=false
for arg in "$@"; do
  [[ "$arg" == "--skip-build" ]] && SKIP_BUILD=true
done

if [[ "$SKIP_BUILD" == true && -d "$INSTALL_DIR/backend/dist" && -d "$INSTALL_DIR/frontend/dist" ]]; then
  echo "Skipping build (--skip-build set and dist directories exist)."
else
  echo "Building..."
  make -C "$INSTALL_DIR" build
fi

# --- Systemd user dir ---
mkdir -p "$SYSTEMD_USER_DIR"

# --- Env file ---
echo "Writing env file: $ENV_FILE"

read_env_var() {
  local key="$1" default="$2"
  local value=""
  if [[ -f "$DOT_ENV" ]]; then
    value=$(grep -E "^${key}=" "$DOT_ENV" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'") || true
  fi
  echo "${value:-$default}"
}

ANTHROPIC_API_KEY=$(read_env_var "ANTHROPIC_API_KEY" "")
CLAUDE_BIN=$(read_env_var "CLAUDE_BIN" "claude")
PORT=$(read_env_var "PORT" "9998")

cat > "$ENV_FILE" <<EOF
ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY
CLAUDE_BIN=$CLAUDE_BIN
PORT=$PORT
EOF

# --- Unit file ---
echo "Installing unit file: $UNIT_FILE"
sed \
  "s|__INSTALL_DIR__|${INSTALL_DIR}|g; s|__NODE_BIN__|${NODE_BIN}|g" \
  "$TEMPLATE" > "$UNIT_FILE"

# --- Enable service ---
systemctl --user daemon-reload
systemctl --user enable --now "$SERVICE_NAME"

# --- Linger ---
echo ""
echo "Enabling linger for $USER..."
echo "(This allows the service to start at boot, even without an active login session.)"
loginctl enable-linger "$USER"

PORT_ACTUAL=$(read_env_var "PORT" "9998")
echo ""
echo "Done! Claude Code Dashboard is running at: http://localhost:${PORT_ACTUAL}"
echo "Check status : systemctl --user status $SERVICE_NAME"
echo "View logs    : journalctl --user -u $SERVICE_NAME -f"
