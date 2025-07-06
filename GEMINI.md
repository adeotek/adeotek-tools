# GEMINI.md

## Repository description

This repository contains a collection of tools written in GO and .NET.

## Tools

### `git-repos-backup`

A tool to backup Git repositories from GitHub or Gitea to a local directory. It supports both public and private repositories, and can be configured to run periodically using a cron job.
The `git-repos-backup` tool is written in Go and provides a simple command-line interface to manage backups of Git repositories.

#### `git-repos-backup` Code Style Guidelines
- Imports: Standard library first, third-party next, grouped with blank lines
- Formatting: Go standard (go fmt), 2-space indentation
- Types: Exported types have comments, use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: Check immediately, descriptive messages, wrap errors with context
- Testing: Standard Go testing package, mock dependencies, separate integration tests
- Documentation: Package and exported function comments follow Go conventions
- Architecture: Follow Go conventions with cmd/, internal/, pkg/ directories

### `sql-migration`

A tool to manage and apply SQL database migrations (SQL scripts). It supports PostgreSQL and SQLite databases and provides a simple command-line interface for managing migrations.
The `sql-migration` tool is written in .NET and allows users to apply to databases migrations (in the form of a collection of SQL scripts).

#### `sql-migration` Code Style Guidelines
- .NET version: .NET 9.0 or later
- Formatting: standard .NET formatting (dotnet format)
- Error handling: prefere using return values instead of exceptions for expected errors, use exceptions for unexpected errors only when absolutely necessary
- Testing: Standard .NET testing package using xUnit, and NSubstitute for mocking
- Documentation: Package and exported function comments follow .NET conventions
- Build: the project should be built with AOT (Ahead of Time) compilation enabled, for both Linux x64 and Windows x64 platforms
- Architecture: Follow .NET conventions with src/, tests/, and tools/ directories
