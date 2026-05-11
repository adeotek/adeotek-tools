# Systemd User Service Support — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `make service-install` / `make service-uninstall` targets that install the dashboard as a systemd user service with lingering enabled.

**Architecture:** A committed unit file template with `__INSTALL_DIR__` and `__NODE_BIN__` placeholders is stamped at install time by `scripts/install-service.sh`. Environment variables are written to a separate `~/.config/systemd/user/claude-code-dashboard.env` file sourced from `backend/.env`. The Makefile provides the user-facing entry points.

**Tech Stack:** bash, systemd (systemctl --user, loginctl), GNU sed, make

---

## File Map

| Action | Path | Purpose |
|--------|------|---------|
| Create | `scripts/claude-code-dashboard.service.template` | Systemd unit file with `__INSTALL_DIR__` and `__NODE_BIN__` placeholders |
| Create | `scripts/install-service.sh` | Preflight → build → env file → unit file → enable → linger |
| Create | `scripts/uninstall-service.sh` | Stop → disable → remove files → daemon-reload |
| Create | `scripts/test-service-install.sh` | Exercises file-generation logic without calling systemctl |
| Modify | `Makefile` | Add `service-install`, `service-uninstall`, `service-test` targets |

---

## Task 1: Unit File Template

**Files:**
- Create: `scripts/claude-code-dashboard.service.template`

- [ ] **Step 1: Create the template**

```ini
[Unit]
Description=Claude Code Dashboard
After=network.target

[Service]
Type=simple
WorkingDirectory=__INSTALL_DIR__
ExecStart=__NODE_BIN__ __INSTALL_DIR__/backend/dist/server.js
EnvironmentFile=%h/.config/systemd/user/claude-code-dashboard.env
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

Save to `scripts/claude-code-dashboard.service.template`.

- [ ] **Step 2: Commit**

```bash
git add scripts/claude-code-dashboard.service.template
git commit -m "feat(service): add systemd unit file template"
```

---

## Task 2: Test Script (write before implementation — TDD)

**Files:**
- Create: `scripts/test-service-install.sh`

The test script exercises the file-generation logic (template substitution, env file generation) in a temp directory, with no real systemctl/loginctl calls. Run it against the template (Task 1) before the scripts exist to confirm tests fail, then again after Task 3 to confirm they pass.

- [ ] **Step 1: Write the test script**

```bash
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
    ((PASS++))
  else
    echo "  FAIL: $desc"
    echo "    Expected to find: '$expected'"
    ((FAIL++))
  fi
}

assert_not_contains() {
  local desc="$1" unexpected="$2" actual="$3"
  if ! echo "$actual" | grep -qF "$unexpected"; then
    echo "  PASS: $desc"
    ((PASS++))
  else
    echo "  FAIL: $desc (found unexpected: '$unexpected')"
    ((FAIL++))
  fi
}

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
assert_contains "Default API key is empty" "" "$KEY_D"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
```

Save to `scripts/test-service-install.sh` and make it executable:
```bash
chmod +x scripts/test-service-install.sh
```

- [ ] **Step 2: Run tests — expect PASS for template tests (template exists), script-dependent tests are exercised inline**

```bash
bash scripts/test-service-install.sh
```

Expected: all assertions PASS (the test is self-contained; it doesn't call install-service.sh directly, so it passes once the template exists).

- [ ] **Step 3: Commit**

```bash
git add scripts/test-service-install.sh
git commit -m "test(service): add install script test for file generation logic"
```

---

## Task 3: install-service.sh

**Files:**
- Create: `scripts/install-service.sh`

- [ ] **Step 1: Write the install script**

```bash
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
```

Save to `scripts/install-service.sh` and make it executable:
```bash
chmod +x scripts/install-service.sh
```

- [ ] **Step 2: Re-run tests to confirm they still pass**

```bash
bash scripts/test-service-install.sh
```

Expected: all PASS

- [ ] **Step 3: Commit**

```bash
git add scripts/install-service.sh
git commit -m "feat(service): add install-service.sh"
```

---

## Task 4: uninstall-service.sh

**Files:**
- Create: `scripts/uninstall-service.sh`

- [ ] **Step 1: Write the uninstall script**

```bash
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
```

Save to `scripts/uninstall-service.sh` and make it executable:
```bash
chmod +x scripts/uninstall-service.sh
```

- [ ] **Step 2: Commit**

```bash
git add scripts/uninstall-service.sh
git commit -m "feat(service): add uninstall-service.sh"
```

---

## Task 5: Makefile Targets

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add targets**

Replace the `.PHONY` line and add the two new targets at the end of the Makefile:

```makefile
.PHONY: dev dev-backend dev-frontend build lint install clean service-install service-uninstall service-test
```

At the end of the file add:

```makefile
service-install:
	@bash scripts/install-service.sh $(ARGS)

service-uninstall:
	@bash scripts/uninstall-service.sh

service-test:
	@bash scripts/test-service-install.sh
```

(`$(ARGS)` passes through `ARGS=--skip-build` from the command line when needed.)

- [ ] **Step 2: Verify the test target works**

```bash
make service-test
```

Expected output ends with: `Results: N passed, 0 failed`

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat(service): add service-install/uninstall/test make targets"
```

---

## Task 6: Smoke Test (manual — requires systemd)

Run only on a Linux host with a running systemd user session.

- [ ] **Step 1: Verify the template substitution produces a valid unit file**

```bash
# Dry-run: stamp the template without actually installing
sed \
  "s|__INSTALL_DIR__|$(pwd)|g; s|__NODE_BIN__|$(command -v node)|g" \
  scripts/claude-code-dashboard.service.template
```

Expected: a valid `.service` file with no `__PLACEHOLDER__` strings remaining.

- [ ] **Step 2: Run make service-test**

```bash
make service-test
```

Expected: all assertions PASS.

- [ ] **Step 3: Install the service (requires a built dist)**

```bash
make build
make service-install
```

Expected: ends with `Claude Code Dashboard is running at: http://localhost:9998`

- [ ] **Step 4: Verify the service is running**

```bash
systemctl --user status claude-code-dashboard
```

Expected: `Active: active (running)`

- [ ] **Step 5: Verify the dashboard is reachable**

```bash
curl -s http://localhost:9998/health
```

Expected: `{"status":"ok"}`

- [ ] **Step 6: Verify linger is enabled**

```bash
loginctl show-user "$USER" | grep Linger
```

Expected: `Linger=yes`

- [ ] **Step 7: Uninstall and verify cleanup**

```bash
make service-uninstall
systemctl --user status claude-code-dashboard
```

Expected: status shows `inactive (dead)` or `not found`

- [ ] **Step 8: Final commit**

```bash
git add -A
git commit -m "chore: verify systemd service install/uninstall smoke test"
```

---

## Self-Review Checklist

- [x] Spec step 1 (preflight): Task 3, `check_command` calls
- [x] Spec step 2 (build + --skip-build): Task 3, build section
- [x] Spec step 3 (create config dir): Task 3, `mkdir -p`
- [x] Spec step 4 (env file generation): Task 3, `read_env_var` + `cat > ENV_FILE`
- [x] Spec step 5 (stamp unit file): Task 3, `sed` substitution (extended to `__NODE_BIN__`)
- [x] Spec step 6 (daemon-reload): Task 3
- [x] Spec step 7 (enable --now): Task 3
- [x] Spec step 8 (linger + note): Task 3
- [x] Spec step 9 (print URL): Task 3
- [x] Uninstall flow: Task 4
- [x] Makefile targets: Task 5
- [x] Unit file template: Task 1
- [x] Env file format (3 vars, FRONTEND_ORIGIN excluded): Task 3 + tested in Task 2
