# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Lint/Test Commands
- Build: `make build`
- Build all platforms: `make build-all`
- Install: `make install`
- Format code: `make fmt`
- Lint code: `make lint`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Run integration tests: `make integration-test` or `RUN_INTEGRATION_TESTS=1 go test -v ./tests`

## Code Style Guidelines
- Imports: Standard library first, third-party next, grouped with blank lines
- Formatting: Go standard (go fmt), 2-space indentation
- Types: Exported types have comments, use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: Check immediately, descriptive messages, wrap errors with context
- Testing: Standard Go testing package, mock dependencies, separate integration tests
- Documentation: Package and exported function comments follow Go conventions
- Architecture: Follow Go conventions with cmd/, internal/, pkg/ directories
