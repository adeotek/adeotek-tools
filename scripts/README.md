# scripts

Standalone utility scripts that don't belong to any specific project.

## bash/

### `pg-backup.sh`

A Bash script for backing up PostgreSQL databases using `pg_dump`. Supports both direct connections and SSH tunnel mode, with include/exclude filtering for selective database backups.

**Dependencies:** `pg_dump`, `psql` (PostgreSQL client tools), `ssh` (for tunnel mode)

**When to use this vs `sql-toolbox`:** Use this script for simple one-off or cron-scheduled backups where you don't need S3 upload, gzip compression, or YAML config management. For production backup pipelines with S3 and config files, use [`sql-toolbox`](../sql-toolbox/).

#### Usage

```bash
# Direct connection
bash/pg-backup.sh -h db.example.com -u postgres --pg-password secret

# SSH tunnel (password via env var)
export PGPASSWORD=secret
bash/pg-backup.sh -s bastion.example.com -i myuser -k ~/.ssh/id_rsa -u postgres

# Selective backup
bash/pg-backup.sh -h db.example.com -u postgres --include "app_db,analytics"
bash/pg-backup.sh -h db.example.com -u postgres --exclude "temp_db,old_db"

# Per-database subdirectories
bash/pg-backup.sh -h db.example.com -u postgres -d /backups --db-subdir
```

Run `bash/pg-backup.sh --help` for the full option reference.

#### Key options

| Option | Default | Description |
|--------|---------|-------------|
| `-h, --host` | — | PostgreSQL host (required for direct connection) |
| `-s, --ssh-host` | — | SSH bastion host (enables tunnel mode) |
| `-u, --user` | — | PostgreSQL username (required) |
| `--pg-password` | `$PGPASSWORD` | Password (or set `PGPASSWORD` env var) |
| `-d, --backup-dir` | `./postgres_backups` | Output directory |
| `--db-subdir` | off | Create a subdirectory per database |
| `--include` | — | Comma-separated list of databases to back up |
| `--exclude` | — | Comma-separated list of databases to skip |
| `-v, --verbose` | off | Verbose `pg_dump` output |

#### Output format

Backups are written as `pg_dump` custom-format (`.dump`) files named `<db>-<YYYYMMDDHHMMSS>.dump`. Restore with:

```bash
pg_restore -h <host> -U <user> -d <target_db> <file>.dump
```
