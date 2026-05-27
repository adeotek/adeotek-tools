# Filter & Configuration Models

> 12 nodes

## Key Concepts

- **FilterRepositories()** (8 connections) — `pkg/filter/filter.go`
- **config.ProviderConfig - Per-provider configuration struct** (5 connections) — `git-repos-backup/internal/config/config.go`
- **integration_test.go** (3 connections) — `tests/integration_test.go`
- **TestIntegration()** (3 connections) — `tests/integration_test.go`
- **testFilterIntegration()** (3 connections) — `tests/integration_test.go`
- **TestFilterRepositories()** (2 connections) — `pkg/filter/filter_test.go`
- **repository.Repository - Common repository struct** (2 connections) — `git-repos-backup/internal/repository/repository.go`
- **README.md - git-repos-backup user documentation** (2 connections) — `git-repos-backup/README.md`
- **filter.go** (1 connections) — `pkg/filter/filter.go`
- **filter_test.go** (1 connections) — `pkg/filter/filter_test.go`
- **config.Config - Top-level configuration struct holding providers slice** (1 connections) — `git-repos-backup/internal/config/config.go`
- **Include-takes-precedence Filter Strategy** (1 connections) — `git-repos-backup/pkg/filter/filter.go`

## Relationships

- [[Git Operations & Mirror Backup]] (3 shared connections)
- [[App Core & CLI Orchestration]] (2 shared connections)
- [[Repository Layer & Exec Injection]] (1 shared connections)

## Source Files

- `git-repos-backup/README.md`
- `git-repos-backup/internal/config/config.go`
- `git-repos-backup/internal/repository/repository.go`
- `git-repos-backup/pkg/filter/filter.go`
- `pkg/filter/filter.go`
- `pkg/filter/filter_test.go`
- `tests/integration_test.go`

## Audit Trail

- EXTRACTED: 25 (78%)
- INFERRED: 7 (22%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*