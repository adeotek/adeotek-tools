# Build, Test & CLI Docs

> 38 nodes

## Key Concepts

- **backup subcommand** (15 connections) — `sql-toolbox/README.md`
- **migration subcommand** (14 connections) — `sql-toolbox/README.md`
- **SQL Toolbox README** (13 connections) — `sql-toolbox/README.md`
- **integration-test.sh** (11 connections) — `integration-test.sh`
- **001_create_users.sql - Create users table** (5 connections) — `sql-toolbox/test-scripts/tables/001_create_users.sql`
- **Migration History Tracking (__migrations_history table)** (4 connections) — `sql-toolbox/README.md`
- **Dry-Run Mode** (4 connections) — `sql-toolbox/README.md`
- **Unified YAML Configuration (migration + backup sections)** (4 connections) — `sql-toolbox/README.md`
- **002_create_posts.sql - Create posts table** (4 connections) — `sql-toolbox/test-scripts/tables/002_create_posts.sql`
- **backup_worker.sh** (3 connections) — `backup_worker.sh`
- **build.sh** (3 connections) — `build.sh`
- **SSH Tunnel Support** (3 connections) — `sql-toolbox/README.md`
- **Connection String Formats (lib/pq, .NET, PostgreSQL URL)** (3 connections) — `sql-toolbox/README.md`
- **S3 Upload Support (AWS S3, MinIO, RustFS)** (3 connections) — `sql-toolbox/README.md`
- **Ordered Script Execution (alphabetical + recursive subdirs)** (3 connections) — `sql-toolbox/README.md`
- **Environment Variable Support (SQLTB_ prefix)** (3 connections) — `sql-toolbox/README.md`
- **Dollar-Quoting SQL Parser ($$, $body$, $inner$, $func$)** (3 connections) — `sql-toolbox/test-scripts/stored_procedures/001_complex_procedure.sql`
- **SQL Toolbox CLAUDE.md** (2 connections) — `sql-toolbox/CLAUDE.md`
- **SHA256 Hash Verification for Script Changes** (2 connections) — `sql-toolbox/README.md`
- **Wildcard Database Backup (database: *)** (2 connections) — `sql-toolbox/README.md`
- **Per-Database S3 Overrides** (2 connections) — `sql-toolbox/README.md`
- **Pure Go PostgreSQL Dump (no pg_dump binary)** (2 connections) — `sql-toolbox/README.md`
- **001_user_posts_view.sql - user_posts view joining users and posts** (2 connections) — `sql-toolbox/test-scripts/views/001_user_posts_view.sql`
- **002_seed_posts.sql - Seed posts data** (2 connections) — `sql-toolbox/test-scripts/data/002_seed_posts.sql`
- **PostgreSQL Database Support** (2 connections) — `sql-toolbox/README.md`
- *... and 13 more nodes in this community*

## Relationships

- No strong cross-community connections detected

## Source Files

- `backup_worker.sh`
- `build.sh`
- `integration-test.sh`
- `sql-toolbox/CLAUDE.md`
- `sql-toolbox/README.md`
- `sql-toolbox/test-scripts/complex_dollar_test.sql`
- `sql-toolbox/test-scripts/data/001_seed_users.sql`
- `sql-toolbox/test-scripts/data/002_seed_posts.sql`
- `sql-toolbox/test-scripts/stored_procedures/001_complex_procedure.sql`
- `sql-toolbox/test-scripts/tables/001_create_users.sql`
- `sql-toolbox/test-scripts/tables/002_create_posts.sql`
- `sql-toolbox/test-scripts/tables/level1/level2/level3/001_deeply_nested.sql`
- `sql-toolbox/test-scripts/views/001_user_posts_view.sql`

## Audit Trail

- EXTRACTED: 114 (89%)
- INFERRED: 14 (11%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*