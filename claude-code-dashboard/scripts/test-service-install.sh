#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE="$SCRIPT_DIR/claude-code-dashboard.service.template"

PASS=0
FAIL=0

assert_contains() {
  local desc="$1" expected="$2" actual="$3"
  if echo "$actual" | grep -qF "$expected"; then
    echo "  PASS: $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $desc"
    echo "    Expected to find: '$expected'"
    FAIL=$((FAIL + 1))
  fi
}

assert_not_contains() {
  local desc="$1" unexpected="$2" actual="$3"
  if ! echo "$actual" | grep -qF "$unexpected"; then
    echo "  PASS: $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $desc (found unexpected: '$unexpected')"
    FAIL=$((FAIL + 1))
  fi
}

if [[ ! -f "$TEMPLATE" ]]; then
  echo "Error: template file not found: $TEMPLATE" >&2
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "=== Unit file template substitution ==="
sed \
  "s|__INSTALL_DIR__|/opt/claude-dashboard|g; s|__NODE_BIN__|/usr/bin/node|g" \
  "$TEMPLATE" > "$TMPDIR/test.service"
UNIT=$(cat "$TMPDIR/test.service")

assert_contains     "WorkingDirectory set"         "WorkingDirectory=/opt/claude-dashboard"                    "$UNIT"
assert_contains     "ExecStart node path"           "ExecStart=/usr/bin/node /opt/claude-dashboard/backend/dist/server.js" "$UNIT"
assert_contains     "EnvironmentFile uses %h"       "EnvironmentFile=%h/.config/systemd/user/claude-code-dashboard.env" "$UNIT"
assert_contains     "Restart=on-failure"            "Restart=on-failure"                                        "$UNIT"
assert_contains     "WantedBy=default.target"       "WantedBy=default.target"                                   "$UNIT"
assert_not_contains "no __INSTALL_DIR__ remaining"  "__INSTALL_DIR__"                                           "$UNIT"
assert_not_contains "no __NODE_BIN__ remaining"     "__NODE_BIN__"                                              "$UNIT"

echo ""
echo "=== Env file generation (with .env) ==="
cat > "$TMPDIR/.env" <<'DOT_ENV'
ANTHROPIC_API_KEY=test-key-abc
CLAUDE_BIN=/home/user/.local/bin/claude
PORT=8888
FRONTEND_ORIGIN=http://localhost:9999
DOT_ENV

read_env_var() {
  local key="$1" default="$2" env_file="$3"
  local value
  value=$(grep -E "^${key}=" "$env_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'") || true
  echo "${value:-$default}"
}

ANTHROPIC_API_KEY=$(read_env_var "ANTHROPIC_API_KEY" ""      "$TMPDIR/.env")
CLAUDE_BIN=$(        read_env_var "CLAUDE_BIN"        "claude" "$TMPDIR/.env")
PORT=$(              read_env_var "PORT"               "9998"   "$TMPDIR/.env")

cat > "$TMPDIR/test.env" <<EOF
ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY
CLAUDE_BIN=$CLAUDE_BIN
PORT=$PORT
EOF

ENV=$(cat "$TMPDIR/test.env")
assert_contains     "ANTHROPIC_API_KEY written"  "ANTHROPIC_API_KEY=test-key-abc"              "$ENV"
assert_contains     "CLAUDE_BIN written"         "CLAUDE_BIN=/home/user/.local/bin/claude"     "$ENV"
assert_contains     "PORT written"               "PORT=8888"                                   "$ENV"
assert_not_contains "FRONTEND_ORIGIN excluded"   "FRONTEND_ORIGIN"                             "$ENV"

echo ""
echo "=== Env file defaults (no .env) ==="
CLAUDE_BIN_D=$(read_env_var "CLAUDE_BIN" "claude" "/nonexistent/.env")
PORT_D=$(      read_env_var "PORT"       "9998"   "/nonexistent/.env")
KEY_D=$(       read_env_var "ANTHROPIC_API_KEY" "" "/nonexistent/.env")

assert_contains "Default CLAUDE_BIN" "claude" "$CLAUDE_BIN_D"
assert_contains "Default PORT"       "9998"   "$PORT_D"

if [[ -z "$KEY_D" ]]; then
  echo "  PASS: Default API key is empty"
  ((PASS++))
else
  echo "  FAIL: Default API key should be empty, got: '$KEY_D'"
  ((FAIL++))
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
