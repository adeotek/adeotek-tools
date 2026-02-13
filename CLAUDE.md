# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository description

This repository contains a collection of tools written in Go.

## Tools

### `git-repos-backup`

A tool to backup Git repositories from GitHub or Gitea to a local directory. It supports both public and private repositories, and can be configured to run periodically using a cron job.
The `git-repos-backup` tool is written in Go and provides a simple command-line interface to manage backups of Git repositories.

#### `git-repos-backup` Build/Lint/Test Commands
- Build: `make build`
- Build all platforms: `make build-all`
- Install: `make install`
- Format code: `make fmt`
- Lint code: `make lint`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Run integration tests: `make integration-test` or `RUN_INTEGRATION_TESTS=1 go test -v ./tests`

#### `git-repos-backup` Code Style Guidelines
- Imports: Standard library first, third-party next, grouped with blank lines
- Formatting: Go standard (go fmt), 2-space indentation
- Types: Exported types have comments, use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: Check immediately, descriptive messages, wrap errors with context
- Testing: Standard Go testing package, mock dependencies, separate integration tests
- Documentation: Package and exported function comments follow Go conventions
- Architecture: Follow Go conventions with cmd/, internal/, pkg/ directories

### `sql-toolbox`

A command-line tool for executing SQL migration scripts and performing database backups for PostgreSQL and SQLite databases.
The `sql-toolbox` tool is written in Go and provides two subcommands: `migration` for managing SQL database migrations and `backup` for multi-database backup with optional gzip compression and S3 upload.

#### `sql-toolbox` Build/Lint/Test Commands
- Build: `make build`
- Build all platforms: `make build-all`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Run integration tests: `bash integration-test.sh`

#### `sql-toolbox` Code Style Guidelines
- Imports: Standard library first, third-party next, grouped with blank lines
- Formatting: Go standard (go fmt)
- Types: Exported types have comments, use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: Check immediately, descriptive messages, wrap errors with context
- Testing: Standard Go testing package, mock dependencies
- Documentation: Package and exported function comments follow Go conventions
- Architecture: Follow Go conventions with cmd/, internal/ directories
