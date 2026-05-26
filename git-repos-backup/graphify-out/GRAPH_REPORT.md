# Graph Report - .  (2026-05-26)

## Corpus Check
- Corpus is ~7,974 words - fits in a single context window. You may not need a graph.

## Summary
- 76 nodes · 113 edges · 8 communities (5 shown, 3 thin omitted)
- Extraction: 81% EXTRACTED · 19% INFERRED · 0% AMBIGUOUS · INFERRED: 21 edges (avg confidence: 0.87)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_App Core & CLI Orchestration|App Core & CLI Orchestration]]
- [[_COMMUNITY_Git Operations & Mirror Backup|Git Operations & Mirror Backup]]
- [[_COMMUNITY_Repository Layer & Exec Injection|Repository Layer & Exec Injection]]
- [[_COMMUNITY_Filter & Configuration Models|Filter & Configuration Models]]
- [[_COMMUNITY_Config Parsing & Providers|Config Parsing & Providers]]
- [[_COMMUNITY_Docker Cron Worker|Docker Cron Worker]]
- [[_COMMUNITY_Provider Type Enum|Provider Type Enum]]
- [[_COMMUNITY_Project Documentation|Project Documentation]]

## God Nodes (most connected - your core abstractions)
1. `Run()` - 19 edges
2. `FetchRepository()` - 10 edges
3. `FilterRepositories()` - 8 edges
4. `GetRepositories()` - 7 edges
5. `CreateFromArgs()` - 5 edges
6. `config.ProviderConfig - Per-provider configuration struct` - 5 edges
7. `git.ExecCommand - Swappable exec.Command variable for testing` - 5 edges
8. `repository.ExecCommand - Swappable exec.Command variable for testing` - 5 edges
9. `Load()` - 4 edges
10. `RunGitFetch()` - 4 edges

## Surprising Connections (you probably didn't know these)
- `README.md - git-repos-backup user documentation` --references--> `Run()`  [INFERRED]
  git-repos-backup/README.md → internal/app/app.go
- `main()` --calls--> `Run()`  [INFERRED]
  main.go → internal/app/app.go
- `main()` --calls--> `Run()`  [INFERRED]
  cmd/git-repos-backup/main.go → internal/app/app.go
- `backup_worker.sh - Docker entrypoint loop script` --references--> `Run()`  [INFERRED]
  git-repos-backup/backup_worker.sh → internal/app/app.go
- `Dual Configuration Mode (YAML file vs CLI args)` --rationale_for--> `Run()`  [INFERRED]
  git-repos-backup/internal/app/app.go → internal/app/app.go

## Hyperedges (group relationships)
- **Core Backup Pipeline: Config -> Fetch Repos -> Filter -> Git Fetch** — app_app_run, repository_repository_getrepositories, filter_filter_filterrepositories, git_git_cli_fetchrepository [EXTRACTED 1.00]
- **Swappable ExecCommand pattern used across git and repository packages for test mocking** — git_git_cli_execcommand, repository_repository_execcommand, concept_swappable_execcommand [INFERRED 0.95]
- **Dual configuration modes: YAML file load vs CLI args construction** — config_config_load, config_config_createfromargs, concept_dual_config_mode [INFERRED 0.95]

## Communities (8 total, 3 thin omitted)

### Community 0 - "App Core & CLI Orchestration"
Cohesion: 0.15
Nodes (15): PrintUsage(), Run(), splitCommaSeparatedList(), TestArgsConfig(), TestConfigLoad(), TestHelpDisplay(), TestPrintUsage(), TestSplitCommaSeparatedList() (+7 more)

### Community 1 - "Git Operations & Mirror Backup"
Cohesion: 0.20
Nodes (13): Bare Repository Mirror Backup Pattern, FetchRepository(), GetGitCommand(), GetRepoPath(), GetRepoUrl(), RepoExists(), RunGitFetch(), RunGitInit() (+5 more)

### Community 2 - "Repository Layer & Exec Injection"
Cohesion: 0.20
Nodes (10): Swappable ExecCommand Pattern for testability, git.ExecCommand - Swappable exec.Command variable for testing, TestFetchRepository(), Repository, repository.ExecCommand - Swappable exec.Command variable for testing, getGiteaRepositories(), getGitHubRepositories(), GetRepositories() (+2 more)

### Community 3 - "Filter & Configuration Models"
Cohesion: 0.20
Nodes (9): Include-takes-precedence Filter Strategy, config.Config - Top-level configuration struct holding providers slice, config.ProviderConfig - Per-provider configuration struct, FilterRepositories(), TestFilterRepositories(), README.md - git-repos-backup user documentation, repository.Repository - Common repository struct, testFilterIntegration() (+1 more)

### Community 4 - "Config Parsing & Providers"
Cohesion: 0.31
Nodes (7): Config, CreateFromArgs(), Load(), TestCreateFromArgs(), TestLoad(), ProviderConfig, ProviderType

## Knowledge Gaps
- **9 isolated node(s):** `backup_worker.sh script`, `ProviderConfig`, `Config`, `Repository`, `main.go - Root convenience wrapper entry point` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Run()` connect `App Core & CLI Orchestration` to `Git Operations & Mirror Backup`, `Repository Layer & Exec Injection`, `Filter & Configuration Models`, `Config Parsing & Providers`?**
  _High betweenness centrality (0.593) - this node is a cross-community bridge._
- **Why does `FetchRepository()` connect `Git Operations & Mirror Backup` to `App Core & CLI Orchestration`, `Repository Layer & Exec Injection`, `Filter & Configuration Models`?**
  _High betweenness centrality (0.318) - this node is a cross-community bridge._
- **Why does `GetRepositories()` connect `Repository Layer & Exec Injection` to `App Core & CLI Orchestration`, `Filter & Configuration Models`?**
  _High betweenness centrality (0.217) - this node is a cross-community bridge._
- **Are the 7 inferred relationships involving `Run()` (e.g. with `main()` and `main()`) actually correct?**
  _`Run()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `FilterRepositories()` (e.g. with `testFilterIntegration()` and `Include-takes-precedence Filter Strategy`) actually correct?**
  _`FilterRepositories()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `backup_worker.sh script`, `ProviderConfig`, `Config` to the rest of the system?**
  _10 weakly-connected nodes found - possible documentation gaps or missing edges._