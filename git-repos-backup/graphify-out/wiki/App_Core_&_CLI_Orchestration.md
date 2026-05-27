# App Core & CLI Orchestration

> 19 nodes

## Key Concepts

- **Run()** (19 connections) — `internal/app/app.go`
- **app_test.go** (6 connections) — `internal/app/app_test.go`
- **app.go** (3 connections) — `internal/app/app.go`
- **splitCommaSeparatedList()** (3 connections) — `internal/app/app.go`
- **PrintUsage()** (3 connections) — `internal/app/app.go`
- **main()** (2 connections) — `main.go`
- **main()** (2 connections) — `cmd/git-repos-backup/main.go`
- **TestPrintUsage()** (2 connections) — `internal/app/app_test.go`
- **TestVersionDisplay()** (2 connections) — `internal/app/app_test.go`
- **TestHelpDisplay()** (2 connections) — `internal/app/app_test.go`
- **TestConfigLoad()** (2 connections) — `internal/app/app_test.go`
- **TestSplitCommaSeparatedList()** (2 connections) — `internal/app/app_test.go`
- **TestArgsConfig()** (2 connections) — `internal/app/app_test.go`
- **backup_worker.sh - Docker entrypoint loop script** (2 connections) — `git-repos-backup/backup_worker.sh`
- **Dual Configuration Mode (YAML file vs CLI args)** (2 connections) — `git-repos-backup/internal/app/app.go`
- **main.go** (1 connections) — `main.go`
- **main.go** (1 connections) — `cmd/git-repos-backup/main.go`
- **main.go - Root convenience wrapper entry point** (1 connections) — `git-repos-backup/main.go`
- **cmd/git-repos-backup/main.go - CLI entry point** (1 connections) — `git-repos-backup/cmd/git-repos-backup/main.go`

## Relationships

- [[Config Parsing & Providers]] (2 shared connections)
- [[Filter & Configuration Models]] (2 shared connections)
- [[Repository Layer & Exec Injection]] (1 shared connections)
- [[Git Operations & Mirror Backup]] (1 shared connections)

## Source Files

- `cmd/git-repos-backup/main.go`
- `git-repos-backup/backup_worker.sh`
- `git-repos-backup/cmd/git-repos-backup/main.go`
- `git-repos-backup/internal/app/app.go`
- `git-repos-backup/main.go`
- `internal/app/app.go`
- `internal/app/app_test.go`
- `main.go`

## Audit Trail

- EXTRACTED: 43 (74%)
- INFERRED: 15 (26%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*