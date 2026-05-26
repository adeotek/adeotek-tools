# Repository Layer & Exec Injection

> 15 nodes

## Key Concepts

- **GetRepositories()** (7 connections) — `internal/repository/repository.go`
- **repository_test.go** (5 connections) — `internal/repository/repository_test.go`
- **git.ExecCommand - Swappable exec.Command variable for testing** (5 connections) — `git-repos-backup/internal/git/git_cli.go`
- **repository.ExecCommand - Swappable exec.Command variable for testing** (5 connections) — `git-repos-backup/internal/repository/repository.go`
- **repository.go** (4 connections) — `internal/repository/repository.go`
- **getGiteaRepositories()** (4 connections) — `internal/repository/repository.go`
- **getGitHubRepositories()** (4 connections) — `internal/repository/repository.go`
- **TestFetchRepository()** (3 connections) — `internal/git/git_cli_test.go`
- **TestGetRepositories_Gitea()** (3 connections) — `internal/repository/repository_test.go`
- **TestGetRepositories_GitHub()** (2 connections) — `internal/repository/repository_test.go`
- **Swappable ExecCommand Pattern for testability** (2 connections) — `git-repos-backup/internal/git/git_cli.go`
- **Repository** (1 connections) — `internal/repository/repository.go`
- **fakeExecCommand()** (1 connections) — `internal/repository/repository_test.go`
- **TestRepository()** (1 connections) — `internal/repository/repository_test.go`
- **TestHelperProcess()** (1 connections) — `internal/repository/repository_test.go`

## Relationships

- [[Git Operations & Mirror Backup]] (4 shared connections)
- [[App Core & CLI Orchestration]] (1 shared connections)
- [[Filter & Configuration Models]] (1 shared connections)

## Source Files

- `git-repos-backup/internal/git/git_cli.go`
- `git-repos-backup/internal/repository/repository.go`
- `internal/git/git_cli_test.go`
- `internal/repository/repository.go`
- `internal/repository/repository_test.go`

## Audit Trail

- EXTRACTED: 40 (83%)
- INFERRED: 8 (17%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*