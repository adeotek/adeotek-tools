# SQL Migration Tool

This is a CLI tool to apply multiple SQL scripts to a database.

## Usage

```
dotnet run --project src/SqlMigration/SqlMigration.csproj -- --scripts-path <path-to-scripts-dir> [options]
```

### Options

- `--scripts-path`: The path to the directory containing the SQL scripts. (Required)
- `--connection-string`: The connection string to the database.
- `--host`: The database host.
- `--port`: The database port.
- `--database`: The database name.
- `--user`: The database user.
- `--password`: The database password.

Connection details can also be provided as environment variables with the `SQL_MIGRATE_` prefix. For example, `SQL_MIGRATE_HOST`, `SQL_MIGRATE_PORT`, etc.

## Building for production

To build the tool with AOT compilation for Linux and Windows, run the following command:

```
dotnet publish src/SqlMigration/SqlMigration.csproj -c Release -r <runtime-identifier>
```

Where `<runtime-identifier>` can be `win-x64` or `linux-x64`.
