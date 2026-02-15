# SQL Toolbox

A command-line tool for executing SQL migration scripts and performing database backups for PostgreSQL and SQLite databases.

## Features

- Support for PostgreSQL and SQLite databases
- **SSH tunnel support** for secure connections through bastion hosts (migration and backup)
- **Multiple connection string formats**: lib/pq, .NET, and PostgreSQL URL formats
- Migration history tracking to avoid re-running scripts
- Automatic database backup before applying migrations
- Database restore from backup
- Multi-database backup via YAML config (`backup` subcommand) with optional gzip compression and S3 upload
- **Per-database S3 bucket and prefix overrides** for flexible backup organization
- Pure Go PostgreSQL dump (no `pg_dump` binary required for `backup`)
- Support for external `pg_dump` binary (optional backup method)
- Dry-run mode for testing migrations without applying changes
- Ordered execution of scripts (tables → views → stored procedures → data)
- SHA256 hash verification to detect script changes
- Environment variable support for configuration
- Unified YAML configuration for both migration and backup
- Verbose logging for debugging

## Installation

### From Source

```bash
git clone https://github.com/adeotek/adeotek-tools.git
cd adeotek-tools/sql-toolbox
make build
```

The binary will be available at `build/sql-toolbox`.

### Using Go Install

```bash
go install github.com/adeotek/adeotek-tools/sql-toolbox/cmd/sql-toolbox@latest
```

## Usage

The tool provides two subcommands: `migration` and `backup`.

### Migration

#### Basic Usage

```bash
sql-toolbox migration --target-path /path/to/sql/scripts --provider postgresql --host localhost --port 5432 --database mydb --user myuser --password mypass
```

#### Using Connection String

```bash
sql-toolbox migration --target-path /path/to/sql/scripts --provider postgresql --connection-string "host=localhost port=5432 dbname=mydb user=myuser password=mypass sslmode=disable"
```

#### SQLite Usage

```bash
sql-toolbox migration --target-path /path/to/sql/scripts --provider sqlite --database /path/to/database.db
```

#### Using YAML Config

```bash
sql-toolbox migration --config config.yaml
```

CLI flags and environment variables override values from the config file.

#### Connection String Formats

The tool supports three connection string formats that are automatically detected and normalized:

1. **lib/pq format** (PostgreSQL driver format):
   ```
   host=localhost port=5432 dbname=mydb user=myuser password=mypass sslmode=disable
   ```

2. **.NET format** (semicolon-separated):
   ```
   Server=localhost;Port=5432;Database=mydb;Username=myuser;Password=mypass;SslMode=Disable;
   ```

3. **PostgreSQL URL format**:
   ```
   postgresql://myuser:mypass@localhost:5432/mydb
   postgres://myuser:mypass@localhost:5432/mydb?sslmode=require
   ```

All formats are automatically converted to lib/pq format internally. Use any format with the `--connection-string` flag or in YAML configuration.

#### SSH Tunnel Support

Connect to databases through SSH bastion hosts for enhanced security. Supports three authentication methods:
- Private key file (`auth_method: "key"`)
- Password (`auth_method: "password"`)
- SSH agent (`auth_method: "agent"`)

Example YAML configuration:

```yaml
migration:
  target_path: "./sql-scripts"
  provider: "postgresql"
  host: "db.internal.example.com"  # Internal database host
  port: 5432
  database: "myapp"
  user: "dbuser"
  password: "dbpass"

  ssh_tunnel:
    enabled: true
    host: "bastion.example.com"
    port: 22
    user: "sshuser"
    auth_method: "key"
    key_file: "~/.ssh/id_rsa"
    local_port: 0  # 0 = auto-assign
```

The tunnel is automatically established before database operations and cleaned up afterward. Works with both `migration` and `backup` commands.

#### Dry Run Mode

```bash
sql-toolbox migration --target-path /path/to/sql/scripts --provider postgresql --host localhost --port 5432 --database mydb --user myuser --password mypass --dry-run
```

#### Verbose Output

```bash
sql-toolbox migration --target-path /path/to/sql/scripts --provider postgresql --host localhost --port 5432 --database mydb --user myuser --password mypass --verbose
```

### Migration Command Line Options

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
| `--config` | `-f` | Path to a YAML configuration file | No |

### Environment Variables

You can use environment variables instead of command-line flags. Prefix them with `SQLTB_`:

- `SQLTB_PROVIDER`
- `SQLTB_CONNECTION_STRING`
- `SQLTB_HOST`
- `SQLTB_PORT`
- `SQLTB_DATABASE`
- `SQLTB_USER`
- `SQLTB_PASSWORD`
- `SQLTB_DRY_RUN`
- `SQLTB_VERBOSE`

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
sql-toolbox migration \
  --target-path ./sql-scripts \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --backup

# SQLite example
sql-toolbox migration \
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
sql-toolbox migration \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --backup-only

# SQLite example
sql-toolbox migration \
  --provider sqlite \
  --database ./mydb.db \
  --backup-only
```

### Restore

Use the `--restore` flag to restore the most recent backup:

```bash
# PostgreSQL example
sql-toolbox migration \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --restore

# SQLite example
sql-toolbox migration \
  --provider sqlite \
  --database ./mydb.db \
  --restore
```

### Dry-run Mode

Both backup and restore operations respect the `--dry-run` flag:

```bash
# Simulate backup without creating actual backup file
sql-toolbox migration \
  --target-path ./sql-scripts \
  --database mydb \
  --backup \
  --dry-run

# Simulate restore without actually restoring
sql-toolbox migration \
  --database mydb \
  --restore \
  --dry-run
```

## Multi-Database Backup (`backup` Subcommand)

The `backup` subcommand provides a YAML-driven multi-database backup workflow for PostgreSQL databases. Unlike the `--backup-only` flag on the migration subcommand (which uses `pg_dump`), this subcommand uses a pure Go implementation and supports backing up multiple databases, gzip compression, and S3/S3-compatible storage upload.

### Quick Start

1. Copy the example config and edit it:
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your database connection details
   ```

2. Run the backup:
   ```bash
   sql-toolbox backup --config config.yaml --verbose
   ```

3. Preview with dry-run:
   ```bash
   sql-toolbox backup --config config.yaml --dry-run
   ```

### Backup Subcommand Flags

| Flag | Short | Description | Required |
|------|-------|-------------|----------|
| `--config` | `-c` | Path to the YAML configuration file | Yes |
| `--verbose` | `-v` | Enable verbose output | No |
| `--dry-run` | `-d` | Simulate execution without making changes | No |

### Unified YAML Configuration

The configuration file supports both `migration` and `backup` sections. Each section is optional.

```yaml
# migration section (optional, for `sql-toolbox migration --config`)
migration:
  target_path: "./sql-scripts"
  provider: "postgresql"
  host: "localhost"
  port: 5432
  database: "myapp"
  user: "admin"
  password: "secret"
  connection_string: ""
  backup: false
  backup_only: false
  restore: false

# backup section (required for `sql-toolbox backup --config`)
backup:
  defaults:
    output_dir: "./backups"
    compress: true
    upload_to_s3: false
    no_owner: true
    clean: true
    schemas: ["public"]

  databases:
    - name: "myapp_prod"
      host: "db.example.com"
      port: 5432
      database: "myapp"
      user: "backup_user"
      password: "secret"
      ssl_mode: "disable"

  s3:
    enabled: false
    bucket: "my-db-backups"
    region: "us-east-1"
    prefix: "sql-toolbox/"
    access_key_id: ""
    secret_access_key: ""
    endpoint: ""
    delete_local_after_upload: false
```

### Configuration Reference

#### `backup.defaults`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `output_dir` | string | `./backups` | Default output directory for dump files |
| `compress` | bool | `true` | Enable gzip compression |
| `upload_to_s3` | bool | `false` | Upload to S3 after dump |
| `no_owner` | bool | `true` | Exclude ownership statements |
| `clean` | bool | `true` | Include `DROP IF EXISTS` before `CREATE` |
| `schemas` | []string | (all) | Schemas to include; empty means all non-system schemas |
| `backup_method` | string | `go` | Backup method: `go` (pure Go) or `pg_dump` (external binary) |
| `ssh_tunnel` | object | | Default SSH tunnel configuration for all databases |

#### `backup.databases[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Identifier for this database (used in file names and S3 keys) |
| `host` | string | Yes* | Database host |
| `port` | int | No | Database port (default: 5432) |
| `database` | string | Yes* | Database name |
| `user` | string | No | Database user |
| `password` | string | No | Database password |
| `ssl_mode` | string | No | SSL mode (default: `disable`) |
| `connection_string` | string | Yes* | Full connection string (lib/pq, .NET, or URL format) |
| `output_dir` | string | No | Per-db output directory override |
| `compress` | bool | No | Per-db compression override |
| `upload_to_s3` | bool | No | Per-db S3 upload override |
| `delete_local_after_upload` | bool | No | Delete local file after S3 upload |
| `schemas` | []string | No | Per-db schema filter override |
| `exclude_tables` | []string | No | Tables to exclude from dump |
| `no_owner` | bool | No | Per-db no-owner override |
| `clean` | bool | No | Per-db clean override |
| `backup_method` | string | No | Per-db backup method override (`go` or `pg_dump`) |
| `ssh_tunnel` | object | No | Per-db SSH tunnel configuration override |
| `s3_prefix` | string | No | Per-db S3 key prefix override (default: from s3 config) |
| `s3_bucket` | string | No | Per-db S3 bucket override (default: from s3 config) |

\* Either `connection_string` or `host` + `database` must be provided.

#### `backup.s3`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable S3 upload globally |
| `bucket` | string | | S3 bucket name (default, can be overridden per database) |
| `region` | string | | AWS region (required for AWS S3, can be any value for MinIO/RustFS) |
| `prefix` | string | | Key prefix for all uploads (default, can be overridden per database) |
| `access_key_id` | string | | AWS access key (or use `AWS_ACCESS_KEY_ID` env var) |
| `secret_access_key` | string | | AWS secret key (or use `AWS_SECRET_ACCESS_KEY` env var) |
| `endpoint` | string | | Custom endpoint for MinIO/RustFS/S3-compatible storage (see below) |
| `delete_local_after_upload` | bool | `false` | Default for deleting local files after upload |

**Endpoint format for MinIO/RustFS:**
- Include protocol and port: `http://localhost:9000` or `https://minio.example.com`
- MinIO example: `http://localhost:9000`
- RustFS example: `http://localhost:8080`
- No trailing slash

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

- **Local files:**
  - Plain SQL: `{name}_backup_{yyyyMMdd_HHmmss}.sql`
  - Compressed: `{name}_backup_{yyyyMMdd_HHmmss}.sql.gz`

- **S3 uploads:**
  - Bucket: Per-database `s3_bucket` or global `bucket` setting
  - Key: `{prefix}{name}/{filename}` where prefix is per-database `s3_prefix` or global `prefix`
  - Example: `production/myapp_prod/myapp_prod_backup_20240115_120000.sql.gz`

### Differences from `--backup-only` Flag

| Feature | `--backup-only` flag | `backup` subcommand |
|---|---|---|
| Database support | Single database | Multiple databases |
| Configuration | CLI flags | YAML config file |
| Dump method | `pg_dump` (external binary) | Pure Go or `pg_dump` (configurable) |
| Compression | No | Optional gzip |
| S3 upload | No | Yes (AWS S3 / MinIO / RustFS) |
| S3 customization | N/A | Per-database bucket/prefix overrides |
| SSH tunnel support | No | Yes (key/password/agent auth) |
| Connection formats | lib/pq only | lib/pq, .NET, PostgreSQL URL |
| SQLite support | Yes | No (PostgreSQL only) |

## Examples

### PostgreSQL Example

```bash
# Using individual connection parameters
sql-toolbox migration \
  --target-path ./sql-scripts \
  --provider postgresql \
  --host localhost \
  --port 5432 \
  --database myapp \
  --user myuser \
  --password mypass \
  --verbose

# Using connection string
sql-toolbox migration \
  --target-path ./sql-scripts \
  --provider postgresql \
  --connection-string "host=localhost port=5432 dbname=myapp user=myuser password=mypass sslmode=disable"
```

### SQLite Example

```bash
sql-toolbox migration \
  --target-path ./sql-scripts \
  --provider sqlite \
  --database ./myapp.db \
  --verbose
```

### Using Environment Variables

```bash
export SQLTB_PROVIDER=postgresql
export SQLTB_HOST=localhost
export SQLTB_PORT=5432
export SQLTB_DATABASE=myapp
export SQLTB_USER=myuser
export SQLTB_PASSWORD=mypass

sql-toolbox migration --target-path <sql-scripts-dir> --verbose
```

### Set environment variables from .env file

```shell
export $(grep -v '^#' .env | xargs)

sql-toolbox migration --target-path <sql-scripts-dir> --verbose
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
   # Format: sql-toolbox/vMAJOR.MINOR.PATCH
   git tag sql-toolbox/v0.7.0
   git push origin sql-toolbox/v0.7.0
   ```

   Or use an annotated tag (recommended):
   ```bash
   git tag -a sql-toolbox/v0.7.0 -m "Release v0.7.0: Description of changes"
   git push origin sql-toolbox/v0.7.0
   ```

3. **Wait for Go proxy indexing:**
   The Go proxy (proxy.golang.org) needs 15-30 minutes to index the new version.

4. **Verify the tag:**
   ```bash
   git ls-remote --tags origin | grep sql-toolbox
   ```

### Installing Specific Versions

Users can install the latest version or a specific version:

```bash
# Install latest version
go install github.com/adeotek/adeotek-tools/sql-toolbox/cmd/sql-toolbox@latest

# Install specific version
go install github.com/adeotek/adeotek-tools/sql-toolbox/cmd/sql-toolbox@v0.7.0
```

### Updating After a New Release

To update to the latest version:

```bash
go install github.com/adeotek/adeotek-tools/sql-toolbox/cmd/sql-toolbox@latest
```

### Important Notes

- **Tag Format:** Must use `sql-toolbox/vX.Y.Z` format (not just `vX.Y.Z`)
- **Version Prefix:** Always prefix tags with the module subdirectory name
- **Semantic Versioning:** Follow semantic versioning (MAJOR.MINOR.PATCH)
- **Go Proxy Cache:** Allow time for proxy.golang.org to index new versions

## License

This project is licensed under the MIT License - see the LICENSE file for details.
