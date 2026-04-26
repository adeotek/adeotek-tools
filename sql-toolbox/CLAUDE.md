# CLAUDE.md — sql-toolbox

A Go CLI tool for executing SQL migrations and performing database backups for PostgreSQL and SQLite. Provides two subcommands: `migration` and `backup`.

## Build / Test Commands

- Build: `make build`
- Build all platforms: `make build-all`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Run integration tests: `bash integration-test.sh`

## Architecture

```
cmd/        # CLI entry points (migration, backup subcommands)
internal/   # Private application logic
```

## Key Features / Gotchas

- **Wildcard backup**: `database: "*"` in config backs up all databases on a server (with exclude list support)
- **Connection string formats**: accepts lib/pq, .NET semicolon-separated, and PostgreSQL URL formats
- **SSH tunnel**: supports connections through bastion hosts
- **PostgreSQL dump**: can use pure Go implementation or external `pg_dump` binary — set appropriately per environment
- **S3 overrides**: per-database bucket/prefix/credentials allow flexible backup organization
- **S3-compatible storage**: works with AWS S3, MinIO, and RustFS

## Code Style

- Imports: standard library first, third-party next, separated by blank lines
- Formatting: `go fmt`
- Types: exported types have doc comments; use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: check immediately; wrap errors with descriptive context
- Testing: standard `testing` package; mock dependencies
- Documentation: package and exported function comments follow Go conventions
