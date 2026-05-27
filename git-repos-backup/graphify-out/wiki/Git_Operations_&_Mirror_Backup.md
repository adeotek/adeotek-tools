# Git Operations & Mirror Backup

> 17 nodes

## Key Concepts

- **FetchRepository()** (10 connections) — `internal/git/git_cli.go`
- **git_cli.go** (7 connections) — `internal/git/git_cli.go`
- **git_cli_test.go** (7 connections) — `internal/git/git_cli_test.go`
- **RunGitFetch()** (4 connections) — `internal/git/git_cli.go`
- **RunGitInit()** (4 connections) — `internal/git/git_cli.go`
- **GetGitCommand()** (4 connections) — `internal/git/git_cli.go`
- **GetRepoPath()** (4 connections) — `internal/git/git_cli.go`
- **RepoExists()** (4 connections) — `internal/git/git_cli.go`
- **GetRepoUrl()** (3 connections) — `internal/git/git_cli.go`
- **TestBasicWorkflow()** (3 connections) — `tests/integration_test.go`
- **TestGetRepoUrl()** (2 connections) — `internal/git/git_cli_test.go`
- **TestGetRepoPath()** (2 connections) — `internal/git/git_cli_test.go`
- **TestRepoExists()** (2 connections) — `internal/git/git_cli_test.go`
- **TestGetGitCommand()** (2 connections) — `internal/git/git_cli_test.go`
- **Bare Repository Mirror Backup Pattern** (2 connections) — `git-repos-backup/internal/git/git_cli.go`
- **fakeExecCommand()** (1 connections) — `internal/git/git_cli_test.go`
- **TestHelperProcess()** (1 connections) — `internal/git/git_cli_test.go`

## Relationships

- [[Repository Layer & Exec Injection]] (4 shared connections)
- [[Filter & Configuration Models]] (3 shared connections)
- [[App Core & CLI Orchestration]] (1 shared connections)

## Source Files

- `git-repos-backup/internal/git/git_cli.go`
- `internal/git/git_cli.go`
- `internal/git/git_cli_test.go`
- `tests/integration_test.go`

## Audit Trail

- EXTRACTED: 52 (84%)
- INFERRED: 10 (16%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*