#!/bin/bash

set -euo pipefail

PG_HOST=""
PG_PORT="5432"
PG_REMOTE_PORT="5432"
PG_USER=""
PG_PASSWORD="${PGPASSWORD:-}"
SSH_HOST=""
SSH_USER=""
SSH_KEY=""
BACKUP_DIR="./postgres_backups"
VV=false  # Verbose mode off by default
NO_COLOR=false
VERBOSE_FLAG=""
USE_SSH_TUNNEL=false
SSH_PID=""
INCLUDE_DBS=""
EXCLUDE_DBS=""

function cecho() {
  local color=$1
  shift

  case $1 in
    -n)
      local args="-n"
      shift
      ;;
    *) local args="";;
  esac

  if [[ "$NO_COLOR" == "true" ]]; then
    echo $args "$@"
    return
  fi
  
  case $color in
    "black") color_code="30";;
    "red") color_code="31";;
    "green") color_code="32";;
    "yellow") color_code="33";;
    "blue") color_code="34";;
    "magenta") color_code="35";;
    "cyan") color_code="36";;
    "white") color_code="37";;
    *) color_code="";;
  esac
  
  if [ -z "$color_code" ]; then
    echo $args "$@"
  else
    echo -e $args "\e[${color_code}m$@\e[0m"
  fi
}

function decho() {
  if [[ "$VV" == "true" ]]; then
    cecho "$@"
  fi
}

function usage() {
    cat << EOF
Usage: $0 [OPTIONS]

OPTIONS:
    -h, --host PG_HOST          PostgreSQL host (required for direct connection)
    -p, --port PG_PORT          Local PostgreSQL port (default: 5432)
    -r, --remote-port PG_PORT   Remote PostgreSQL port when using SSH tunnel (default: 5432)
    -u, --user PG_USER          PostgreSQL username (required)
    --pg-password PASSWORD      PostgreSQL password (optional if PGPASSWORD env var is set)
    -s, --ssh-host SSH_HOST     SSH server hostname/IP (enables SSH tunnel mode)
    -i, --ssh-user SSH_USER     SSH username (required for SSH tunnel)
    -k, --ssh-key SSH_KEY       Path to SSH private key (required for SSH tunnel)
    -d, --backup-dir DIR        Local backup directory (default: ./postgres_backups)
    --include DB1,DB2,DB3       Only backup specified databases (comma-separated)
    --exclude DB1,DB2,DB3       Exclude specified databases (comma-separated)
    -v, --verbose               Enable verbose output
    -n, --no-color              Disable colored output
    --help                      Show this help message

EXAMPLES:
    # SSH tunnel mode
    $0 -s server.com -i myuser -k ~/.ssh/id_rsa -u postgres
    
    # Direct connection
    $0 -h db.server.com -u postgres --pg-password mypass
    
    # Only backup specific databases
    $0 -h db.server.com -u postgres --include "myapp,myapp_test"
    
    # Exclude system/temp databases
    $0 -s server.com -i user -k ~/.ssh/key -u postgres --exclude "temp_db,old_db"
    
    # Using environment variable for password
    export PGPASSWORD=mypass
    $0 -s server.com -i user -k ~/.ssh/key -u postgres

EOF
}

function cleanup() {
    decho "white" "Checking if SSH tunnel is opened (PID: $SSH_PID)..."
    if [[ -n "$SSH_PID" ]] && kill -0 "$SSH_PID" 2>/dev/null; then
        cecho "cyan" "Closing SSH tunnel (PID: $SSH_PID)..."
        kill "$SSH_PID"
        wait "$SSH_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            PG_HOST="$2"
            shift 2
            ;;
        -p|--port)
            PG_PORT="$2"
            shift 2
            ;;
        -r|--remote-port)
            PG_REMOTE_PORT="$2"
            shift 2
            ;;
        -u|--user)
            PG_USER="$2"
            shift 2
            ;;
        --pg-password)
            PG_PASSWORD="$2"
            shift 2
            ;;
        -s|--ssh-host)
            SSH_HOST="$2"
            shift 2
            ;;
        -i|--ssh-user)
            SSH_USER="$2"
            shift 2
            ;;
        -k|--ssh-key)
            SSH_KEY="$2"
            shift 2
            ;;
        -d|--backup-dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        --include)
            INCLUDE_DBS="$2"
            shift 2
            ;;
        --exclude)
            EXCLUDE_DBS="$2"
            shift 2
            ;;
        -v|--verbose)
            VV=true
            VERBOSE_FLAG="--verbose"
            shift
            ;;
        -n|--no-color)
            NO_COLOR=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if [[ -n "$SSH_HOST" ]]; then
    if [[ -z "$SSH_USER" || -z "$SSH_KEY" ]]; then
        cecho "red" "Error: SSH tunnel requires --ssh-user and --ssh-key parameters" >&2
        usage >&2
        exit 1
    fi
    USE_SSH_TUNNEL=true
    PG_HOST="localhost"
else
    if [[ -z "$PG_HOST" ]]; then
        cecho "red" "Error: Direct connection requires --host parameter" >&2
        usage >&2
        exit 1
    fi
fi

if [[ -z "$PG_USER" ]]; then
    cecho "red" "Error: --pg-user parameter is required" >&2
    usage >&2
    exit 1
fi

if [[ -z "$PG_PASSWORD" ]]; then
    cecho "red" "Error: PostgreSQL password must be provided via PGPASSWORD environment variable or --pg-password argument" >&2
    exit 1
fi

if [[ "$USE_SSH_TUNNEL" == "true" ]] && [[ ! -f "$SSH_KEY" ]]; then
    cecho "red" "Error: SSH key file not found: $SSH_KEY" >&2
    exit 1
fi

if ! command -v pg_dump >/dev/null 2>&1; then
    cecho "red" "Error: pg_dump not found. Please install PostgreSQL client tools." >&2
    exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
    cecho "red" "Error: psql not found. Please install PostgreSQL client tools." >&2
    exit 1
fi

cecho "cyan" "Creating backup directory: $BACKUP_DIR"
mkdir -p "$BACKUP_DIR"

export PGPASSWORD="$PG_PASSWORD"

if [[ "$USE_SSH_TUNNEL" == "true" ]]; then
    cecho "cyan" "Establishing SSH tunnel to $SSH_HOST..."
    
    # Start SSH tunnel in background and capture PID
    ssh -N -L "$PG_PORT:$PG_HOST:$PG_REMOTE_PORT" \
        -i "$SSH_KEY" \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        -o LogLevel=ERROR \
        "$SSH_USER@$SSH_HOST" & SSH_PID=$!

    # Wait a moment for SSH to establish or fail
    sleep 2
    
    # Check if SSH process is still running
    if ! kill -0 "$SSH_PID" 2>/dev/null; then
        cecho "red" "Error: Failed to establish SSH tunnel. Port $PG_PORT may be in use." >&2
        exit 1
    fi
    
    cecho "yellow" "SSH tunnel established (PID: $SSH_PID)"
else
    cecho "cyan" "Using direct connection to PostgreSQL..."
fi

cecho "cyan" "Testing connection to PostgreSQL..."
if ! psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c "SELECT version();" >/dev/null 2>&1; then
    if [[ "$USE_SSH_TUNNEL" == "true" ]]; then
        cecho "red" "Error: Cannot connect to PostgreSQL through SSH tunnel" >&2
    else
        cecho "red" "Error: Cannot connect to PostgreSQL directly" >&2
    fi
    exit 1
fi

cecho "cyan" "Querying user databases..."
USER_DATABASES=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -t -c \
    "SELECT datname FROM pg_database WHERE datistemplate = false AND datname NOT IN ('postgres');" | \
    grep -v '^$' | sed 's/^ *//')

if [[ -z "$USER_DATABASES" ]]; then
    cecho "blue" "No user databases found"
    exit 0
fi

# Apply include/exclude filtering
FILTERED_DATABASES=""
for db in $USER_DATABASES; do
    SKIP_DB=false
    
    # Check exclude list
    if [[ -n "$EXCLUDE_DBS" ]]; then
        IFS=',' read -ra EXCLUDE_ARRAY <<< "$EXCLUDE_DBS"
        for exclude_db in "${EXCLUDE_ARRAY[@]}"; do
            if [[ "$db" == "$exclude_db" ]]; then
                decho "yellow" "Excluding database: $db"
                SKIP_DB=true
                break
            fi
        done
    fi
    
    # Check include list (if specified, only include listed databases)
    if [[ -n "$INCLUDE_DBS" ]] && [[ "$SKIP_DB" == "false" ]]; then
        INCLUDE_MATCH=false
        IFS=',' read -ra INCLUDE_ARRAY <<< "$INCLUDE_DBS"
        for include_db in "${INCLUDE_ARRAY[@]}"; do
            if [[ "$db" == "$include_db" ]]; then
                INCLUDE_MATCH=true
                break
            fi
        done
        if [[ "$INCLUDE_MATCH" == "false" ]]; then
            decho "yellow" "Not in include list, skipping database: $db"
            SKIP_DB=true
        fi
    fi
    
    if [[ "$SKIP_DB" == "false" ]]; then
        FILTERED_DATABASES="$FILTERED_DATABASES $db"
    fi
done

# Trim leading whitespace
FILTERED_DATABASES=$(echo "$FILTERED_DATABASES" | sed 's/^ *//')

if [[ -z "$FILTERED_DATABASES" ]]; then
    cecho "blue" "No databases match the include/exclude criteria"
    exit 0
fi

cecho "yellow" "Found $(echo "$USER_DATABASES" | wc -w) databases, backing up $(echo "$FILTERED_DATABASES" | wc -w):"
TIMESTAMP=$(date +"%Y%m%d%H%M%S")

decho "cyan" "Starting backup process..."
for db in $FILTERED_DATABASES; do
    cecho "magenta" "-> Backing up database: $db"
    BACKUP_FILE="$BACKUP_DIR/${db}-${TIMESTAMP}.dump"

    if pg_dump -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$db" \
        --no-password \
        --format=custom \
        --no-owner \
        --no-privileges \
        $VERBOSE_FLAG \
        -f "$BACKUP_FILE"; then
        
        cecho "yellow" "✓ Successfully backed up $db to $BACKUP_FILE"
    else
        cecho "red" "✗ Failed to backup database: $db" >&2
    fi
done

cecho "green" "Backup process completed!"
cecho "green" "Backups saved to: $BACKUP_DIR"
