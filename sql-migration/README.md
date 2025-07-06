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

The tool expects SQL scripts to be organized in the following directory structure for optimal execution order:

```
sql-scripts/
├── tables/
│   ├── 001_create_users.sql
│   └── 002_create_posts.sql
├── views/
│   └── 001_user_posts_view.sql
├── stored_procedures/
│   └── 001_get_user_posts.sql
└── data/
    └── 001_seed_data.sql
```

Scripts are executed in this order:
1. `tables/` - Database table creation scripts
2. `views/` - Database view creation scripts
3. `stored_procedures/` - Stored procedure creation scripts
4. `data/` - Data insertion/seeding scripts

Within each directory, scripts are executed in alphabetical order.

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

## License

This project is licensed under the MIT License - see the LICENSE file for details.
