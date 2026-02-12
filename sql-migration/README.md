# SQL Migration Tool

A command-line tool for executing SQL migration scripts against PostgreSQL and SQLite databases with migration history tracking.

## Features

- Support for PostgreSQL and SQLite databases
- Migration history tracking to avoid re-running scripts
- Automatic database backup before applying migrations
- Database restore from backup
- Multi-database backup via YAML config (`backup-only` subcommand) with optional gzip compression and S3 upload
- Pure Go PostgreSQL dump (no `pg_dump` binary required for `backup-only`)
- Dry-run mode for testing migrations without applying changes
- Ordered execution of scripts (tables → views → stored procedures → data)
- SHA256 hash verification to detect script changes
- Environment variable support for configuration
- Verbose logging for debugging

## Installation

### From Source

```bash
git clone https://github.com/adeotek/adeotek-tools.git
cd adeotek-tools/sql-migration
make build
```

The binary will be available at `build/sql-migration`.

### Using Go Install

```bash
go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@latest
```

## Usage

### Basic Usage

```bash
sql-migration --target-path /path/to/sql/scripts --provider postgresql --host localhost --port 5432 --database mydb --user myuser --password mypass
```

### Using Connection String

```bash
sql-migration --target-path /path/to/sql/scripts --provider postgresql --connection-string "host=localhost port=5432 dbname=mydb user=myuser password=mypass sslmode=disable"
```

### SQLite Usage

```bash
sql-migration --target-path /path/to/sql/scripts --provider sqlite --database /path/to/database.db
```

### Dry Run Mode

```bash
sql-migration --target-path /path/to/sql/scripts --provider postgresql --host localhost --port 5432 --database mydb --user myuser --password mypass --dry-run
```

### Verbose Output

```bash
sql-migration --target-path /path/to/sql/scripts --provider postgresql --host localhost --port 5432 --database mydb --user myuser --password mypass --verbose
```

## Command Line Options

| Flag | Short | Description | Required |
|------|-------|-------------|----------|
| `--target-path` | `-t` | Path to the SQL scripts directory | Yes (except for --restore and --backup-only) |
| `--provider` | `-r` | Database provider (postgresql/sqlite) | No (default: postgresql) |
| `--connection-string` | `-c` | Database connection string | No |
| `--host` | `-o` | Database host | No |
| `--port` | `-p` | Database port | No |
| `--database` | `-b` | Database name | No |
| `--user` | `-u` | Database user | No |
| `--password` | `-s` | Database password | No |
| `--dry-run` | `-d` | Run in dry-run mode | No |
| `--verbose` | `-v` | Enable verbose output | No |
| `--backup` | | Backup database before applying migrations | No |
| `--backup-only` | | Create database backup without running migrations | No |
| `--restore` | | Restore last database backup and skip migrations | No |
| `--version` | | Show version information | No |

## Environment Variables

You can use environment variables instead of command-line flags. Prefix them with `CLI_SQL_MIGRATION_`:

- `CLI_SQL_MIGRATION_PROVIDER`
- `CLI_SQL_MIGRATION_CONNECTION_STRING`
- `CLI_SQL_MIGRATION_HOST`
- `CLI_SQL_MIGRATION_PORT`
- `CLI_SQL_MIGRATION_DATABASE`
- `CLI_SQL_MIGRATION_USER`
- `CLI_SQL_MIGRATION_PASSWORD`
- `CLI_SQL_MIGRATION_DRY_RUN`
- `CLI_SQL_MIGRATION_VERBOSE`

## Script Organization

Within the target directory, scripts are executed in alphabetical order, and then any subdirectories are processed in the same manner in alphabetical order recursively.
To achieve the correct result, the tool expects SQL scripts to be organized in the optimal execution order.

### Example Directory Structure

```
sql-scripts/
├── 01_tables/
│   ├── 01_audit/
│   │   └── 002_create_audit.sql
│   ├── 02_custom/
│   │   ├── 001_create_custom1.sql
│   │   └── 002_create_custom2.sql
│   ├── 001_create_users.sql
│   └── 002_create_posts.sql
├── 02_views/
│   └── 001_user_posts_view.sql
├── 03_stored_procedures/
│   └── 001_get_user_posts.sql
└── 04_data/
│   ├── 01_post_seed/
│   │   └── 001_seed_posts.sql
    └── 001_seed_data.sql
```

The above scripts will be executed in the following order:

1. `01_tables/001_create_users.sql`
2. `01_tables/002_create_posts.sql`
3. `01_tables/01_audit/002_create_audit.sql`
4. `01_tables/02_custom/001_create_custom1.sql`
5. `01_tables/02_custom/002_create_custom2.sql`
6. `02_views/001_user_posts_view.sql`
7. `03_stored_procedures/001_get_user_posts.sql`
8. `04_data/001_seed_data.sql`
9. `04_data/01_post_seed/001_seed_posts.sql`


## Migration History

The tool creates a `__migrations_history` table to track executed scripts. This table stores:
- Script file name
- SHA256 hash of the script content
- Execution timestamp

Scripts are only re-executed if their content has changed (detected by hash comparison).

## Backup and Restore

The tool supports automatic database backup before applying migrations and manual restore of previous backups.

### Backup

Use the `--backup` flag to create a backup before applying migrations:

```bash
# PostgreSQL example
sql-migration \
  --target-path ./sql-scripts \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --backup

# SQLite example
sql-migration \
  --target-path ./sql-scripts \
  --provider sqlite \
  --database ./mydb.db \
  --backup
```

**Backup behavior:**
- Backups are only created if there are unapplied migration scripts
- Backup files are stored in `.db-backups` directory in the current working directory
- Backup filenames include a timestamp: `{database}_backup_{yyyyMMdd_HHmmss}.{ext}`
- PostgreSQL backups use `pg_dump` and create `.sql` files
- SQLite backups are simple file copies with `.db` extension
- If backup fails, migrations will not be applied

**Prerequisites for PostgreSQL:**
- `pg_dump` must be installed and accessible in PATH
- User must have sufficient privileges to dump the database

### Backup Only

Use the `--backup-only` flag to create a backup without running any migrations:

```bash
# PostgreSQL example
sql-migration \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --backup-only

# SQLite example
sql-migration \
  --provider sqlite \
  --database ./mydb.db \
  --backup-only
```

**Backup-only behavior:**
- Creates a backup regardless of whether there are unapplied scripts
- No migrations are run when `--backup-only` flag is used
- The `--target-path` flag is not required for backup-only operations
- Uses the same backup mechanism as `--backup` flag (pg_dump for PostgreSQL, file copy for SQLite)
- Backup files are stored in `.db-backups` directory with timestamps

### Restore

Use the `--restore` flag to restore the most recent backup:

```bash
# PostgreSQL example
sql-migration \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --restore

# SQLite example
sql-migration \
  --provider sqlite \
  --database ./mydb.db \
  --restore
```

**Restore behavior:**
- Restores the most recent backup from `.db-backups` directory
- No migrations are run when `--restore` flag is used
- The `--target-path` flag is not required for restore operations
- PostgreSQL restore uses `psql` command
- SQLite restore overwrites the database file with the backup copy

**Prerequisites for PostgreSQL:**
- `psql` must be installed and accessible in PATH
- User must have sufficient privileges to restore the database

### Dry-run Mode

Both backup and restore operations respect the `--dry-run` flag:

```bash
# Simulate backup without creating actual backup file
sql-migration \
  --target-path ./sql-scripts \
  --database mydb \
  --backup \
  --dry-run

# Simulate restore without actually restoring
sql-migration \
  --database mydb \
  --restore \
  --dry-run
```

## Multi-Database Backup (`backup-only` Subcommand)

The `backup-only` subcommand provides a YAML-driven multi-database backup workflow for PostgreSQL databases. Unlike the `--backup-only` flag on the root command (which uses `pg_dump`), this subcommand uses a pure Go implementation and supports backing up multiple databases, gzip compression, and S3/S3-compatible storage upload.

### Quick Start

1. Copy the example config and edit it:
   ```bash
   cp backup.yaml.example backup.yaml
   # Edit backup.yaml with your database connection details
   ```

2. Run the backup:
   ```bash
   sql-migration backup-only --config backup.yaml --verbose
   ```

3. Preview with dry-run:
   ```bash
   sql-migration backup-only --config backup.yaml --dry-run
   ```

### Subcommand Flags

| Flag | Short | Description | Required |
|------|-------|-------------|----------|
| `--config` | `-c` | Path to the backup YAML configuration file | Yes |
| `--verbose` | `-v` | Enable verbose output | No |
| `--dry-run` | `-d` | Simulate execution without making changes | No |

### YAML Configuration

The configuration file defines default settings, one or more database targets, and optional S3 upload settings.

```yaml
defaults:
  output_dir: "./backups"
  compress: true              # gzip compression, enabled by default
  upload_to_s3: false         # upload to S3 if s3 is configured, per-db overridable
  no_owner: true              # exclude ownership statements from dump
  clean: true                 # include DROP IF EXISTS before CREATE
  schemas: ["public"]         # omit or leave empty for full backup (all schemas)

databases:
  - name: "myapp_prod"
    host: "db.example.com"
    port: 5432
    database: "myapp"
    user: "backup_user"
    password: "secret"
    ssl_mode: "disable"
    output_dir: "./backups/prod"      # per-db override
    compress: true
    upload_to_s3: true                # per-db override
    delete_local_after_upload: true   # per-db override (default: from s3 config)
    schemas: ["public", "audit"]
    exclude_tables: ["sessions", "cache"]

  - name: "myapp_staging"
    connection_string: "host=staging port=5432 dbname=myapp user=admin password=pass sslmode=disable"
    upload_to_s3: false               # skip S3 upload for this database

s3:
  enabled: false
  bucket: "my-db-backups"
  region: "us-east-1"
  prefix: "sql-migration/"
  access_key_id: ""            # or use AWS_ACCESS_KEY_ID env var
  secret_access_key: ""
  endpoint: ""                 # for MinIO/S3-compatible storage
  delete_local_after_upload: false
```

### Configuration Reference

#### `defaults`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `output_dir` | string | `./backups` | Default output directory for dump files |
| `compress` | bool | `true` | Enable gzip compression |
| `upload_to_s3` | bool | `false` | Upload to S3 after dump |
| `no_owner` | bool | `true` | Exclude ownership statements |
| `clean` | bool | `true` | Include `DROP IF EXISTS` before `CREATE` |
| `schemas` | []string | (all) | Schemas to include; empty means all non-system schemas |

#### `databases[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Identifier for this database (used in file names and S3 keys) |
| `host` | string | Yes* | Database host |
| `port` | int | No | Database port (default: 5432) |
| `database` | string | Yes* | Database name |
| `user` | string | No | Database user |
| `password` | string | No | Database password |
| `ssl_mode` | string | No | SSL mode (default: `disable`) |
| `connection_string` | string | Yes* | Full connection string (alternative to individual params) |
| `output_dir` | string | No | Per-db output directory override |
| `compress` | bool | No | Per-db compression override |
| `upload_to_s3` | bool | No | Per-db S3 upload override |
| `delete_local_after_upload` | bool | No | Delete local file after S3 upload (falls back to `s3.delete_local_after_upload`) |
| `schemas` | []string | No | Per-db schema filter override |
| `exclude_tables` | []string | No | Tables to exclude from dump |
| `no_owner` | bool | No | Per-db no-owner override |
| `clean` | bool | No | Per-db clean override |

\* Either `connection_string` or `host` + `database` must be provided.

#### `s3`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable S3 upload globally |
| `bucket` | string | | S3 bucket name (required when enabled) |
| `region` | string | | AWS region |
| `prefix` | string | | Key prefix for all uploads |
| `access_key_id` | string | | AWS access key (or use `AWS_ACCESS_KEY_ID` env var) |
| `secret_access_key` | string | | AWS secret key (or use `AWS_SECRET_ACCESS_KEY` env var) |
| `endpoint` | string | | Custom endpoint for MinIO/S3-compatible storage |
| `delete_local_after_upload` | bool | `false` | Default for deleting local files after upload |

### What Gets Dumped

The pure Go dump produces SQL output covering:

- Extensions
- Schemas
- Enum types
- Sequences
- Tables (columns, defaults, NOT NULL, primary keys, unique constraints)
- Views
- Functions
- Table data (as INSERT statements)
- Indexes (post-data)
- Foreign keys (post-data)
- Triggers (post-data)

### Backup Output

- Plain SQL files: `{name}_backup_{yyyyMMdd_HHmmss}.sql`
- Compressed files: `{name}_backup_{yyyyMMdd_HHmmss}.sql.gz`
- S3 keys: `{prefix}{name}/{filename}`

### Differences from `--backup-only` Flag

| | `--backup-only` flag | `backup-only` subcommand |
|---|---|---|
| Database support | Single database | Multiple databases |
| Configuration | CLI flags | YAML config file |
| Dump method | Shells out to `pg_dump` | Pure Go (no external binary) |
| Compression | No | Optional gzip |
| S3 upload | No | Yes (AWS S3 / MinIO) |
| SQLite support | Yes | No (PostgreSQL only) |

## Examples

### PostgreSQL Example

```bash
# Using individual connection parameters
sql-migration \
  --target-path ./sql-scripts \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database myapp \
  --user myuser \
  --password mypass \
  --verbose

# Using connection string
sql-migration \
  --target-path ./sql-scripts \
  --provider postgresql \
  --connection-string "host=localhost port=5432 dbname=myapp user=myuser password=mypass sslmode=disable"
```

### SQLite Example

```bash
sql-migration \
  --target-path ./sql-scripts \
  --provider sqlite \
  --database ./myapp.db \
  --verbose
```

### Using Environment Variables

```bash
export CLI_SQL_MIGRATION_PROVIDER=postgresql
export CLI_SQL_MIGRATION_HOST=localhost
export CLI_SQL_MIGRATION_PORT=5432
export CLI_SQL_MIGRATION_DATABASE=myapp
export CLI_SQL_MIGRATION_USER=myuser
export CLI_SQL_MIGRATION_PASSWORD=mypass

sql-migration --target-path <sql-scripts-dir> --verbose
```

### Set environment variables from .env file

```shell
export $(grep -v '^#' .env | xargs)

sql-migration --target-path <sql-scripts-dir> --verbose
```

## Building

### Build for Current Platform

```bash
make build
```

### Build for All Platforms

```bash
make build-all
```

### Run Tests

```bash
make test
```

### Clean Build Artifacts

```bash
make clean
```

## Publishing a New Version

This tool is part of a monorepo, so version tags must be prefixed with the module path.

### Steps to Publish

1. **Commit all changes:**
   ```bash
   git add .
   git commit -m "Your commit message"
   git push origin main
   ```

2. **Create and push a version tag:**
   ```bash
   # Format: sql-migration/vMAJOR.MINOR.PATCH
   git tag sql-migration/v0.4.0
   git push origin sql-migration/v0.4.0
   ```

   Or use an annotated tag (recommended):
   ```bash
   git tag -a sql-migration/v0.4.0 -m "Release v0.4.0: Description of changes"
   git push origin sql-migration/v0.4.0
   ```

3. **Wait for Go proxy indexing:**
   The Go proxy (proxy.golang.org) needs 15-30 minutes to index the new version.

4. **Verify the tag:**
   ```bash
   git ls-remote --tags origin | grep sql-migration
   ```

### Installing Specific Versions

Users can install the latest version or a specific version:

```bash
# Install latest version
go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@latest

# Install specific version
go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@v0.4.0
```

### Updating After a New Release

To update to the latest version:

```bash
go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@latest
```

If the new version doesn't appear immediately, try:

```bash
# Clear module cache
go clean -modcache
go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@latest

# Or bypass proxy cache
GOPROXY=direct go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@latest
```

On Windows PowerShell:
```powershell
$env:GOPROXY="direct"; go install github.com/adeotek/adeotek-tools/sql-migration/cmd/sql-migration@latest
```

### Important Notes

- **Tag Format:** Must use `sql-migration/vX.Y.Z` format (not just `vX.Y.Z`)
- **Version Prefix:** Always prefix tags with the module subdirectory name
- **Semantic Versioning:** Follow semantic versioning (MAJOR.MINOR.PATCH)
- **Go Proxy Cache:** Allow time for proxy.golang.org to index new versions

## License

This project is licensed under the MIT License - see the LICENSE file for details.
