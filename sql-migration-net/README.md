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

## Building for production

To build the tool with AOT compilation for Linux and Windows, run the following command:

```
dotnet publish src/SqlMigration/SqlMigration.csproj -c Release -r <runtime-identifier>
```

Where `<runtime-identifier>` can be `win-x64` or `linux-x64`.
