# AGENTS: sql-toolbox

Purpose: Help AI coding agents be productive when working on the `sql-toolbox` tool.

Quick actions
- Build: `make build` or `./build.sh`
- Run tests: `make test` (or `go test -v ./...`)
- Run integration tests: `bash integration-test.sh`
- Format: `make fmt`
- Lint: `make lint`

Project layout (important entry points)
- CLI entry: [cmd/sql-toolbox/](sql-toolbox/cmd/sql-toolbox/)
- App logic: [internal/app/](sql-toolbox/internal/app/)
- Database layer: [internal/database/](sql-toolbox/internal/database/)
- Models & services: [internal/models/](sql-toolbox/internal/models/) and [internal/services/](sql-toolbox/internal/services/)
- Test fixtures: [test-scripts/](sql-toolbox/test-scripts/)
- Integration helpers: [integration-test.sh](sql-toolbox/integration-test.sh)

Key conventions and notes for agents
- Prefer Makefile targets and `integration-test.sh` rather than ad-hoc commands. See [Makefile](sql-toolbox/Makefile) and [integration-test.sh](sql-toolbox/integration-test.sh).
- Integration tests exercise SQLite migrations and use dry-run; do not assume a live DB unless tests explicitly set one up.
- Config: use `config.yaml.example` and `config.env.example` as authoritative sources for env var names and defaults.
- Environment variables often use the `SQLTB_` prefix; prefer CLI flags over env vars when writing automation to match existing patterns.
- Preserve error wrapping and migration order semantics when changing migration code — migration order matters.

Useful links
- Guide / README: [README.md](sql-toolbox/README.md)
- Config examples: [config.yaml.example](sql-toolbox/config.yaml.example) and [config.env.example](sql-toolbox/config.env.example)
- CLAUDE notes: [CLAUDE.md](sql-toolbox/CLAUDE.md)

When to create additional customization files
- If we need package-specific automation (e.g., a test-runner agent or an integration-test CI hook), create a skill under `.vscode` or a repo-level `skills/` entry and reference this AGENTS file.

If you want more CI-level guidance, commit hooks, or an automated integration-test skill, tell me which area to expand.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, invoke the `skill` tool with `skill: "graphify"` before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
