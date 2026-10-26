# Graph Report - git-repos-backup  (2026-10-26)

## Corpus Check
- 22 files · ~12,761 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 5 file(s) not represented in the graph (top: (none) 4, .example 1)

## Summary
- 156 nodes · 411 edges · 11 communities (7 shown, 4 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 33 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ee3d85fb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Git Repos Backup
- git_cli.go
- testing.T
- Run
- repository.go
- os/exec.Cmd
- CLAUDE.md — git-repos-backup
- AGENTS: git-repos-backup
- backup_worker.sh
- github.com/adeotek/adeotek-tools/git-repos-backup

## God Nodes (most connected - your core abstractions)
1. `Run()` - 16 edges
2. `ProviderConfig` - 16 edges
3. `FetchRepository()` - 14 edges
4. `GetRepositories()` - 14 edges
5. `Git Repos Backup` - 11 edges
6. `GetGitCommandForRemote()` - 9 edges
7. `Repository` - 7 edges
8. `RunGitFetch()` - 6 edges
9. `GetGitCommand()` - 6 edges
10. `assertCredentialHelper()` - 6 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Run()`  [EXTRACTED]
  cmd/git-repos-backup/main.go → internal/app/app.go
- `Run()` --calls--> `FilterRepositories()`  [EXTRACTED]
  internal/app/app.go → pkg/filter/filter.go
- `main()` --calls--> `Run()`  [EXTRACTED]
  main.go → internal/app/app.go
- `FilterRepositories()` --references--> `ProviderConfig`  [EXTRACTED]
  pkg/filter/filter.go → internal/config/config.go
- `TestFetchRepositoryRealGit()` --calls--> `FetchRepository()`  [EXTRACTED]
  tests/fetch_integration_test.go → internal/git/git_cli.go

## Import Cycles
- None detected.

## Communities (11 total, 4 thin omitted)

### Community 0 - "Git Repos Backup"
Cohesion: 0.08
Nodes (25): 1. Using a configuration file, 2. Using command-line arguments, Build, Building the Docker image locally, Code Quality, Command-line Options, Configuration, Development (+17 more)

### Community 1 - "git_cli.go"
Cohesion: 0.14
Nodes (24): ProviderConfig, EnsureHead(), FetchRepository(), GetGitCommand(), GetGitCommandForRemote(), GetRepoPath(), GetRepoUrl(), isScpLikeUrl() (+16 more)

### Community 3 - "testing.T"
Cohesion: 0.14
Nodes (22): TestHelperProcess(), recordingExecCommand(), TestGetRepositories_ForgejoIsSupported(), TestGetRepositories_UnknownProvider(), TestRunCurl_GitHubUsesBearer(), writeTempFile(), GetRepositories(), TestGetGiteaRepositories_Pagination() (+14 more)

### Community 4 - "Run"
Cohesion: 0.11
Nodes (21): main(), Config, ProviderType, PrintUsage(), Run(), splitCommaSeparatedList(), TestArgsConfig(), TestConfigLoad() (+13 more)

### Community 5 - "repository.go"
Cohesion: 0.40
Nodes (8): appendRepositories(), getGiteaCompatibleRepositories(), getGitHubRepositories(), Repository, TestRedactedCurlCommand(), redactedCurlCommand(), runCurl(), apiRepository

### Community 6 - "os/exec.Cmd"
Cohesion: 0.24
Nodes (9): envValue(), fakeExecCommand(), hasArg(), hasEnv(), TestFetchRepositoryStripsURLCredentialsForAllProviders(), TestGetGitCommandForRemote(), TestRunGitFetchUsesCredentialHelper(), pagedExecCommand() (+1 more)

### Community 7 - "CLAUDE.md — git-repos-backup"
Cohesion: 0.33
Nodes (5): Architecture, Build / Lint / Test Commands, CLAUDE.md — git-repos-backup, Code Style, graphify

## Knowledge Gaps
- **1 isolated node(s):** `github.com/adeotek/adeotek-tools/git-repos-backup`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 36 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Run()` connect `Run` to `git_cli.go`, `git_cli_test.go`, `testing.T`?**
  _High betweenness centrality (0.038) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `Run()` (e.g. with `TestArgsConfig()` and `TestConfigLoad()`) actually correct?**
  _`Run()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/adeotek/adeotek-tools/git-repos-backup` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Git Repos Backup` be split into smaller, more focused modules?**
  _Cohesion score 0.07692307692307693 - nodes in this community are weakly interconnected._
- **Why does `FetchRepository()` connect `git_cli.go` to `testing.T`, `Run`, `repository.go`, `os/exec.Cmd`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Are the 2 inferred relationships involving `FetchRepository()` (e.g. with `TestFetchRepository()` and `TestFetchRepositoryStripsURLCredentialsForAllProviders()`) actually correct?**
  _`FetchRepository()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Should `git_cli.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13675213675213677 - nodes in this community are weakly interconnected._