# CLAUDE.md — git-repos-backup

A Go CLI tool to backup Git repositories from GitHub or Gitea to a local directory. Supports public and private repositories; designed to run on a schedule via cron.

## Build / Lint / Test Commands

- Build: `make build`
- Build all platforms: `make build-all`
- Install: `make install`
- Format: `make fmt`
- Lint: `make lint`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Run integration tests: `make integration-test` or `RUN_INTEGRATION_TESTS=1 go test -v ./tests`

## Architecture

Follows standard Go project layout:

```
cmd/        # CLI entry points
internal/   # Private application logic
pkg/        # Exported/shared packages
```

## Code Style

- Imports: standard library first, third-party next, separated by blank lines
- Formatting: `go fmt` (2-space indentation)
- Types: exported types have doc comments; use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: check immediately; wrap errors with descriptive context
- Testing: standard `testing` package; mock dependencies; integration tests live in `tests/`
- Documentation: package and exported function comments follow Go conventions

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
