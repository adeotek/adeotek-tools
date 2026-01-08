# SQL Migration Tool

This is a CLI tool to apply multiple SQL scripts to a database.

## Usage

```
dotnet run --project src/SqlMigration/SqlMigration.csproj -- --target-path <path-to-scripts-dir> [options]
```

### Options

- `--target-path|-t`: The path to the directory containing the SQL scripts. (Required)
- `--connection-string|-c`: The connection string to the database.
- `--provider|-r`: The database provider (PostgreSQL or SQLite, default PostgreSQL).
- `--host|-h`: The database host.
- `--port|-p`: The database port.
- `--database|-b`: The database name.
- `--user|-u`: The database user.
- `--password|-s`: The database password.
- `--dry-run|-d`: Run the command in dry-run mode, which simulates the execution without making any changes.
- `--verbose|-v`: Enable verbose output, providing detailed information about the command execution.
- `--backup`: Backup the database before applying migrations (only if there are unapplied scripts). Backups are stored in `.sql-migration-backups` directory.
- `--restore`: Restore the last database backup and skip running migrations.

Connection details can also be provided as environment variables with the `CLI_SQL_MIGRATION_` prefix. For example, `CLI_SQL_MIGRATION_HOST`, `CLI_SQL_MIGRATION_PORT`, etc.

### Set environment variables from .env file

```shell
export $(grep -v '^#' .env | xargs)
```

## Backup and Restore

The tool supports automatic database backup before applying migrations and manual restore of previous backups.

### Backup

Use the `--backup` flag to create a backup before applying migrations:

```bash
# PostgreSQL example
dotnet run --project src/SqlMigration/SqlMigration.csproj -- \
  --target-path ./scripts \
  --provider PostgreSQL \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --backup

# SQLite example
dotnet run --project src/SqlMigration/SqlMigration.csproj -- \
  --target-path ./scripts \
  --provider SQLite \
  --database ./mydb.db \
  --backup
```

**Backup behavior:**
- Backups are only created if there are unapplied migration scripts
- Backup files are stored in `.sql-migration-backups` directory in the current working directory
- Backup filenames include a timestamp: `{database}_backup_{yyyyMMdd_HHmmss}.{ext}`
- PostgreSQL backups use `pg_dump` and create `.sql` files
- SQLite backups are simple file copies with `.db` extension
- If backup fails, migrations will not be applied

**Prerequisites for PostgreSQL:**
- `pg_dump` must be installed and accessible in PATH
- User must have sufficient privileges to dump the database

### Restore

Use the `--restore` flag to restore the most recent backup:

```bash
# PostgreSQL example
dotnet run --project src/SqlMigration/SqlMigration.csproj -- \
  --provider PostgreSQL \
  --host localhost \
  --port 5432 \
  --database mydb \
  --user myuser \
  --password mypass \
  --restore

# SQLite example
dotnet run --project src/SqlMigration/SqlMigration.csproj -- \
  --provider SQLite \
  --database ./mydb.db \
  --restore
```

**Restore behavior:**
- Restores the most recent backup from `.sql-migration-backups` directory
- No migrations are run when `--restore` flag is used
- The `--target-path` flag is not required for restore operations
- PostgreSQL restore uses `psql` command
- SQLite restore overwrites the database file with the backup copy

**Prerequisites for PostgreSQL:**
- `psql` must be installed and accessible in PATH
- User must have sufficient privileges to restore the database

### Dry-run mode

Both backup and restore operations respect the `--dry-run` flag:

```bash
# Simulate backup without creating actual backup file
dotnet run --project src/SqlMigration/SqlMigration.csproj -- \
  --target-path ./scripts \
  --database mydb \
  --backup \
  --dry-run

# Simulate restore without actually restoring
dotnet run --project src/SqlMigration/SqlMigration.csproj -- \
  --database mydb \
  --restore \
  --dry-run
```

## Building for production

To build the tool with AOT compilation for Linux and Windows, run the following command:

```
dotnet publish src/SqlMigration/SqlMigration.csproj -c Release -r <runtime-identifier>
```

Where `<runtime-identifier>` can be `win-x64` or `linux-x64`.
