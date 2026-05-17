# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository structure

This is a **monorepo** containing multiple independent, unrelated projects in different languages and tech stacks. Projects do not share code or dependencies.

## Navigation for agents

Each project has its own root directory and its own `CLAUDE.md` file with project-specific build commands, code style, and architecture notes. **Always read the project's own `CLAUDE.md` before working on it.**

## Projects

| Project | Path | Stack | Description |
|---------|------|-------|-------------|
| `git-migration` | `git-migration/` | Go | Migrate repos from Gitea to Forgejo with full metadata |
| `git-repos-backup` | `git-repos-backup/` | Go | Backup GitHub/Gitea repos to local storage |
| `sql-toolbox` | `sql-toolbox/` | Go | SQL migrations and PostgreSQL/SQLite backup with S3 support |
| `homelab-chatbot` | `homelab-chatbot/` | Python (FastAPI) + Next.js | Homelab AI chatbot with backend and frontend |
| `bnr-rates-extractor` | `bnr-rates-extractor/` | .NET 10 (C#) | CLI to extract BNR exchange rates to CSV |
| `docker-net-tools` | `docker-net-tools/` | Docker | Utility Docker networking tools |

## Shared conventions

- Projects are independent — changes in one do not affect others
- Each project manages its own dependencies, build system, and tests
- CI is handled per-project via `.github/workflows/`
