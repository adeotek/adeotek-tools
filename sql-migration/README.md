# SQL Migration Tool

A command-line tool for executing SQL migration scripts against PostgreSQL and SQLite databases with migration history tracking.

## Features

- Support for PostgreSQL and SQLite databases
- Migration history tracking to avoid re-running scripts
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
| `--target-path` | `-t` | Path to the SQL scripts directory | Yes |
| `--provider` | `-r` | Database provider (postgresql/sqlite) | No (default: postgresql) |
| `--connection-string` | `-c` | Database connection string | No |
| `--host` | `-o` | Database host | No |
| `--port` | `-p` | Database port | No |
| `--database` | `-b` | Database name | No |
| `--user` | `-u` | Database user | No |
| `--password` | `-s` | Database password | No |
| `--dry-run` | `-d` | Run in dry-run mode | No |
| `--verbose` | `-v` | Enable verbose output | No |
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
