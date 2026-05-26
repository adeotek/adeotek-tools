# AGENTS: git-repos-backup

Purpose: Help AI coding agents be productive when working on the `git-repos-backup` tool.

Quick actions
- Build: `make build`
- Run tests: `make test` (or `go test -v ./...`)
- Run integration tests: `RUN_INTEGRATION_TESTS=1 make integration-test`
- Format: `make fmt`
- Lint: `make lint`

Project layout (important entry points)
- CLI entry: [cmd/git-repos-backup/](git-repos-backup/cmd/git-repos-backup/)
- App logic: [internal/app/](git-repos-backup/internal/app/)
- Config: [internal/config/](git-repos-backup/internal/config/)
- Git helpers: [internal/git/](git-repos-backup/internal/git/)
- Repo models: [internal/repository/](git-repos-backup/internal/repository/)
- Public utilities: [pkg/filter/](git-repos-backup/pkg/filter/)
- Integration tests: [tests/](git-repos-backup/tests/)

Key conventions and notes for agents
- Follow existing Makefile targets rather than inventing ad-hoc commands. See [Makefile](git-repos-backup/Makefile).
- Use `go fmt` and `make fmt` before committing; repository follows Go formatting rules.
- Tests: respect the `RUN_INTEGRATION_TESTS=1` guard when running integration tests. See [tests/integration_test.go](git-repos-backup/tests/integration_test.go).
- Configuration: prefer reading or linking to `config.yaml.example` instead of duplicating content. See [config.yaml.example](git-repos-backup/config.yaml.example).
- Error handling and logging follow repository patterns (see code under `internal/`). Preserve error wrapping and context when refactoring.

Useful links
- Guide / README: [README.md](git-repos-backup/README.md)
- CLAUDE / workspace notes: [CLAUDE.md](git-repos-backup/CLAUDE.md)

When to create additional customization files
- If we need package-specific automation (e.g., a test-runner agent or an integration-test hook), create a skill under `.vscode` or a repo-level `skills/` entry and reference this AGENTS file.

If anything here is unclear or you want extra rules (commit checks, CI details, or automatic test harnesses), tell me which area to expand.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, invoke the `skill` tool with `skill: "graphify"` before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
