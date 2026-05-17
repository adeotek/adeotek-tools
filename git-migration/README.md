# git-migration

Migrate repositories from a self-hosted [Gitea](https://gitea.io) instance to a self-hosted [Forgejo](https://forgejo.org) instance.

Uses Forgejo's built-in migration API to transfer full repository metadata server-side: git history, issues, pull requests, labels, milestones, releases, and wikis.

## Installation

```bash
git clone https://github.com/adeotek/adeotek-tools
cd adeotek-tools/git-migration
make build
# binary: ./git-migration
```

Or install directly to `$GOPATH/bin/`:

```bash
cd adeotek-tools/git-migration && make install
# then use: git-migration
```

## Authentication

Set tokens via environment variables or CLI flags (flags take priority):

```bash
export GITEA_TOKEN=your-gitea-token
export FORGEJO_TOKEN=your-forgejo-token
```

Or pass them directly:

```bash
git-migration --gitea-token abc123 --forgejo-token xyz789 ...
```

## Usage

```
git-migration [flags]

Flags:
  -exclude string       Comma-separated repo names or full names to exclude
  -filter string        Glob pattern to filter repo names (e.g. infra-*)
  -forgejo-token string Forgejo API token (or set FORGEJO_TOKEN)
  -forgejo-url string   Forgejo server URL (required)
  -gitea-token string   Gitea API token (or set GITEA_TOKEN)
  -gitea-url string     Gitea server URL (required)
  -help                 Print usage and exit
  -map-org value        Map source org to destination org (format: src:dst, repeatable)
  -on-conflict string   Behaviour when repo already exists: skip|fail|remigrate (default: skip)
  -orgs string          Comma-separated Gitea orgs to migrate
  -users string         Comma-separated Gitea users to migrate
  -verbose              Show verbose output
  -version              Print version and exit
  -dry-run              Print migration plan without making changes
```

## Examples

### Dry run: preview what would be migrated

```bash
git-migration \
  --gitea-url https://gitea.lan \
  --forgejo-url https://forgejo.lan \
  --orgs myorg \
  --dry-run
```

### Migrate an org with org name mapping

```bash
git-migration \
  --gitea-url https://gitea.lan \
  --forgejo-url https://forgejo.lan \
  --orgs old-org \
  --map-org old-org:new-org
```

### Migrate only repos matching a pattern, excluding one

```bash
git-migration \
  --gitea-url https://gitea.lan \
  --forgejo-url https://forgejo.lan \
  --orgs myorg \
  --filter 'infra-*' \
  --exclude infra-legacy
```

### Re-migrate everything, overwriting existing repos

```bash
git-migration \
  --gitea-url https://gitea.lan \
  --forgejo-url https://forgejo.lan \
  --orgs myorg \
  --on-conflict remigrate
```

### Migrate all repos from specific users

```bash
git-migration \
  --gitea-url https://gitea.lan \
  --forgejo-url https://forgejo.lan \
  --users user1,user2
```

### Migrate all visible repos (requires broad token permissions)

```bash
git-migration \
  --gitea-url https://gitea.lan \
  --forgejo-url https://forgejo.lan
```

## Notes

- **Migration is server-side**: Forgejo pulls from Gitea directly. The tool only orchestrates API calls.
- **Migration is asynchronous**: the tool logs success when the request is accepted (HTTP 201). Check the Forgejo UI for final status.
- If neither `--orgs` nor `--users` is specified, all repos visible to the Gitea token are enumerated.
- Destination orgs that don't exist in Forgejo are created automatically.
- Exit code is non-zero if any repo migration failed.

## Building from Source

### Prerequisites

- Go 1.18 or later

### Build

```bash
make build          # Single platform (linux amd64)
make build-all      # All platforms (linux, windows, darwin amd64)
make test           # Run tests
make fmt            # Format code
make lint           # Run linter (go vet)
make clean          # Remove binaries
```

## Contributing

See the main [adeotek-tools](https://github.com/adeotek/adeotek-tools) repository for contribution guidelines.
