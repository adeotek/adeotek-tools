# CLAUDE.md — git-migration

A Go CLI tool to migrate repositories from a self-hosted Gitea instance to a self-hosted Forgejo instance. Supports full metadata migration (issues, PRs, labels, milestones, releases, wiki) via Forgejo's server-side migration API.

## Build / Lint / Test Commands

- Build: `make build` (creates `git-migration` in current dir)
- Build all platforms: `make build-all` (linux, windows, darwin amd64; output in `bin/`)
- Install: `make install` (copies binary to `$GOPATH/bin/`)
- Format: `make fmt`
- Lint: `go vet ./...` or `make lint`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`

## Architecture

```
cmd/git-migration/    Entry point; imports internal/app
internal/
  ├── app/            CLI entry point (Run, PrintUsage), wires all packages
  ├── config/         Config struct, OrgMappings type, env var resolution, validation
  ├── gitea/          Gitea REST client — paginated repo enumeration (org, user, or all)
  ├── forgejo/        Forgejo REST client — repo-exists check, org creation, migration trigger
  └── migration/      Orchestrator — filter, org mapping, conflict handling, migration loop
```

Zero external dependencies. All HTTP via stdlib `net/http`.

## Code Style

- Imports: standard library first, then internal packages, separated by blank lines
- Formatting: `go fmt`
- Error handling: check immediately; wrap errors with `fmt.Errorf("context: %w", err)`
- Testing: `net/http/httptest` for mocking API endpoints; tests in same package as source

## Key Concepts

- **Org Mappings** (`--map-org src:dst`): Repeatable flag. Maps source Gitea orgs to destination Forgejo orgs.
- **Filtering** (`--filter 'infra-*'`): Glob pattern to filter repo names (shell-style wildcards).
- **Exclusions** (`--exclude repo1,repo2`): Comma-separated list of repo names or full names (`org/repo`) to skip.
- **Conflict Handling** (`--on-conflict skip|fail|remigrate`): Behavior when repo already exists in Forgejo.
- **Dry Run** (`--dry-run`): Shows plan without making any API calls to Forgejo.
- **Environment Tokens**: `GITEA_TOKEN` and `FORGEJO_TOKEN` env vars are fallbacks for CLI flags.

## Testing Strategy

- **Unit tests**: Config validation, org mapping logic, filtering/exclusion logic
- **Integration tests**: Mock Gitea and Forgejo API responses using `httptest`; verify client methods and orchestrator flow
- Tests live in `_test.go` files next to their source; run with `make test`

## API Patterns

Both Gitea and Forgejo clients use:
- Pagination via query params (`?page=N&limit=50`)
- Bearer token auth: `Authorization: token <token>`
- JSON request/response bodies
- HTTP status codes: 200/201 for success, 404 for not found, 409 for conflict
