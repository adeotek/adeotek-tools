#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Fedora / RHEL: include system CA bundle so Anthropic API calls succeed
CA_BUNDLE="/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"
[[ -f "$CA_BUNDLE" ]] && export NODE_EXTRA_CA_CERTS="$CA_BUNDLE"

# Bootstrap .env on first run
if [[ ! -f "$ROOT_DIR/backend/.env" ]]; then
  echo "backend/.env not found — copying from .env.example"
  cp "$ROOT_DIR/.env.example" "$ROOT_DIR/backend/.env"
fi

# Load env vars (sets PORT, CLAUDE_BIN, ANTHROPIC_API_KEY, etc.)
set -a
# shellcheck source=/dev/null
source "$ROOT_DIR/backend/.env"
set +a

echo "Installing dependencies..."
cd "$ROOT_DIR/frontend" && npm install
cd "$ROOT_DIR/backend"  && npm install

echo "Building..."
cd "$ROOT_DIR/frontend" && npm run build
cd "$ROOT_DIR/backend"  && npm run build

PORT="${PORT:-9998}"
echo "Starting Claude Code Dashboard → http://localhost:${PORT}"
exec node "$ROOT_DIR/backend/dist/server.js"
