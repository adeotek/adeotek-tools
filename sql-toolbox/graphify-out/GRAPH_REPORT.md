# Graph Report - sql-toolbox  (2026-11-02)

## Corpus Check
- 54 files · ~33,531 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 6 file(s) not represented in the graph (top: (none) 4, .example 2)

## Summary
- 470 nodes · 1287 edges · 30 communities (20 shown, 10 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 174 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `538d2047`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- testing.T
- PgDumpService
- s3_upload_service_test.go
- runMigration
- Migration
- SQL Toolbox
- connection.go
- Multi-Database Backup (`backup` Subcommand)
- CLAUDE.md — sql-toolbox
- DatabaseTarget
- integration-test.sh
- SSHTunnel
- ConnectionParameters
- Backup and Restore
- Building
- Examples
- Publishing a New Version
- Configuration Reference
- AGENTS: sql-toolbox
- backup_worker.sh
- build.sh
- github.com/adeotek/adeotek-tools/sql-toolbox
- MigrationHistoryRepository
- users

## God Nodes (most connected - your core abstractions)
1. `PgDumpService` - 27 edges
2. `DatabaseTarget` - 25 edges
3. `ConnectionParameters` - 24 edges
4. `SSHTunnel` - 19 edges
5. `writeTempConfig()` - 18 edges
6. `LoadBackupConfig()` - 14 edges
7. `BackupService` - 14 edges
8. `NewBackupService()` - 14 edges
9. `NewDatabaseDumper()` - 13 edges
10. `NewSqlScriptsHelpers()` - 13 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Run()`  [EXTRACTED]
  cmd/sql-toolbox/main.go → internal/app/app.go
- `main()` --calls--> `Run()`  [EXTRACTED]
  main.go → internal/app/app.go
- `NewMigrationService()` --calls--> `NewSqlScriptsHelpers()`  [INFERRED]
  internal/services/migration_service.go → internal/services/sql_scripts_helpers.go
- `Run()` --calls--> `newBackupCmd()`  [INFERRED]
  internal/app/app.go → internal/app/backup_cmd.go
- `Run()` --calls--> `newMigrationCmd()`  [INFERRED]
  internal/app/app.go → internal/app/migration_cmd.go

## Import Cycles
- None detected.

## Communities (30 total, 10 thin omitted)

### Community 0 - "testing.T"
Cohesion: 0.06
Nodes (71): ExpandWildcardDatabasesWithConnStr(), LoadBackupConfig(), boolPtr(), containsAny(), strPtr(), TestBackupConfig_ExpandWildcards_NoWildcards(), TestBackupConfig_Validate(), TestDatabaseTarget_GetEffectiveBackupMethod() (+63 more)

### Community 2 - "PgDumpService"
Cohesion: 0.15
Nodes (14): escapeStringLiteral(), formatInsertValue(), orderViews(), parsePostgresArray(), qualifiedName(), quoteIdentifier(), quoteLiteral(), TestEscapeStringLiteral() (+6 more)

### Community 3 - "s3_upload_service_test.go"
Cohesion: 0.38
Nodes (5): BuildS3Key(), TestBuildS3Key(), TestMockS3Uploader(), MockS3Uploader, mockUploadCall

### Community 4 - "runMigration"
Cohesion: 0.12
Nodes (15): viperKey, main(), Run(), newBackupCmd(), runBackup(), getViper(), setViper(), newMigrationCmd() (+7 more)

### Community 5 - "Migration"
Cohesion: 0.17
Nodes (12): Basic Usage, Connection String Formats, Dry Run Mode, Environment Variables, Migration, Migration Command Line Options, SQLite Usage, SSH Tunnel Support (+4 more)

### Community 6 - "SQL Toolbox"
Cohesion: 0.20
Nodes (9): Example Directory Structure, Features, From Source, Installation, License, Migration History, Script Organization, SQL Toolbox (+1 more)

### Community 7 - "connection.go"
Cohesion: 0.11
Nodes (23): activeTunnelLocalPort(), buildConnInfo(), convertDotNetFormat(), convertPostgreSQLURL(), DatabaseProvider, normalizeConnectionString(), ParseConnectionParameters(), parseConnInfo() (+15 more)

### Community 8 - "Multi-Database Backup (`backup` Subcommand)"
Cohesion: 0.25
Nodes (8): Backup Output, Backup Subcommand Flags, Differences from `--backup-only` Flag, Multi-Database Backup (`backup` Subcommand), Quick Start, Unified YAML Configuration, What Gets Dumped, Wildcard Database Backup

### Community 9 - "CLAUDE.md — sql-toolbox"
Cohesion: 0.29
Nodes (6): Architecture, Build / Test Commands, CLAUDE.md — sql-toolbox, Code Style, graphify, Key Features / Gotchas

### Community 10 - "DatabaseTarget"
Cohesion: 0.06
Nodes (25): BackupConfig, DatabaseTarget, S3Config, QueryDatabases(), validateBackupMethod(), TestBackupConfig_Validate_RejectsPartialS3Credentials(), TestBackupConfig_Validate_RejectsUnknownBackupMethod(), createOutputWriter() (+17 more)

### Community 11 - "integration-test.sh"
Cohesion: 0.40
Nodes (4): integration-test.sh script, SQLTB_DATABASE, SQLTB_PROVIDER, SQLTB_VERBOSE

### Community 12 - "SSHTunnel"
Cohesion: 0.12
Nodes (5): SSHTunnelConfig, ExpandWildcardDatabasesWithTunnel(), EstablishIfNeeded(), NewSSHTunnel(), SSHTunnel

### Community 13 - "ConnectionParameters"
Cohesion: 0.07
Nodes (30): Factory, NewFactory(), ConnectionParameters, NewMigrationHistoryRepository(), decompressBackup(), isSQLiteDatabase(), sanitizeFileName(), vacuumSQLiteInto() (+22 more)

### Community 14 - "Backup and Restore"
Cohesion: 0.40
Nodes (5): Backup, Backup and Restore, Backup Only, Dry-run Mode, Restore

### Community 15 - "Building"
Cohesion: 0.40
Nodes (5): Build for All Platforms, Build for Current Platform, Building, Clean Build Artifacts, Run Tests

### Community 16 - "Examples"
Cohesion: 0.40
Nodes (5): Examples, PostgreSQL Example, Set environment variables from .env file, SQLite Example, Using Environment Variables

### Community 17 - "Publishing a New Version"
Cohesion: 0.40
Nodes (5): Important Notes, Installing Specific Versions, Publishing a New Version, Steps to Publish, Updating After a New Release

### Community 18 - "Configuration Reference"
Cohesion: 0.50
Nodes (4): `backup.databases[]`, `backup.defaults`, `backup.s3`, Configuration Reference

### Community 23 - "MigrationHistoryRepository"
Cohesion: 0.23
Nodes (4): ScriptExecutionHistory, upsertExecutedScript(), MigrationHistoryRepository, upsertQuerier

### Community 24 - "users"
Cohesion: 0.40
Nodes (3): users, posts, user_posts

## Knowledge Gaps
- **4 isolated node(s):** `github.com/adeotek/adeotek-tools/sql-toolbox`, `SQLTB_PROVIDER`, `SQLTB_DATABASE`, `SQLTB_VERBOSE`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 113 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `SSHTunnel` connect `SSHTunnel` to `migration_cmd.go`?**
  _High betweenness centrality (0.056) - this node is a cross-community bridge._
- **What connects `github.com/adeotek/adeotek-tools/sql-toolbox`, `SQLTB_PROVIDER`, `SQLTB_DATABASE` to the rest of the system?**
  _4 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.060939060939060936 - nodes in this community are weakly interconnected._
- **Why does `DatabaseTarget` connect `DatabaseTarget` to `testing.T`, `migration_cmd.go`, `SSHTunnel`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Should `PgDumpService` be split into smaller, more focused modules?**
  _Cohesion score 0.14878048780487804 - nodes in this community are weakly interconnected._
- **Why does `PgDumpService` connect `PgDumpService` to `migration_cmd.go`, `ConnectionParameters`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Should `runMigration` be split into smaller, more focused modules?**
  _Cohesion score 0.11692307692307692 - nodes in this community are weakly interconnected._