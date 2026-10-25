# Graph Report - sql-toolbox  (2026-10-22)

## Corpus Check
- 52 files · ~28,533 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 6 file(s) not represented in the graph (top: (none) 4, .example 2)

## Summary
- 408 nodes · 1095 edges · 23 communities (18 shown, 5 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 151 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `3a89ce3f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- testing.T
- PgDumpService
- s3_upload_service_test.go
- runMigration
- Migration
- SQL Toolbox
- NewDatabaseDumper
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

## God Nodes (most connected - your core abstractions)
1. `DatabaseTarget` - 25 edges
2. `PgDumpService` - 25 edges
3. `ConnectionParameters` - 23 edges
4. `SSHTunnel` - 17 edges
5. `writeTempConfig()` - 15 edges
6. `BackupService` - 14 edges
7. `NewBackupService()` - 13 edges
8. `NewDatabaseDumper()` - 13 edges
9. `BackupConfig` - 12 edges
10. `LoadConfig()` - 12 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Run()`  [EXTRACTED]
  cmd/sql-toolbox/main.go → internal/app/app.go
- `main()` --calls--> `Run()`  [EXTRACTED]
  main.go → internal/app/app.go
- `TestExpandWildcardDatabases_EmptyExcludeDb()` --calls--> `ExpandWildcardDatabasesWithConnStr()`  [INFERRED]
  internal/models/backup_config_test.go → internal/models/backup_config.go
- `NewMigrationService()` --calls--> `NewSqlScriptsHelpers()`  [INFERRED]
  internal/services/migration_service.go → internal/services/sql_scripts_helpers.go
- `TestQualifiedName()` --calls--> `qualifiedName()`  [INFERRED]
  internal/services/pg_dump_service_test.go → internal/services/pg_dump_service.go

## Import Cycles
- None detected.

## Communities (23 total, 5 thin omitted)

### Community 0 - "testing.T"
Cohesion: 0.06
Nodes (66): LoadBackupConfig(), boolPtr(), containsAny(), strPtr(), TestDatabaseTarget_GetEffectiveBackupMethod(), TestDatabaseTarget_GetEffectiveCompress(), TestDatabaseTarget_GetEffectiveDeleteLocalAfterUpload(), TestDatabaseTarget_GetEffectiveOutputDir() (+58 more)

### Community 2 - "PgDumpService"
Cohesion: 0.35
Nodes (3): qualifiedName(), quoteIdentifier(), PgDumpService

### Community 3 - "s3_upload_service_test.go"
Cohesion: 0.38
Nodes (5): BuildS3Key(), TestBuildS3Key(), TestMockS3Uploader(), MockS3Uploader, mockUploadCall

### Community 4 - "runMigration"
Cohesion: 0.13
Nodes (15): viperKey, main(), Run(), newBackupCmd(), runBackup(), getViper(), setViper(), newMigrationCmd() (+7 more)

### Community 5 - "Migration"
Cohesion: 0.17
Nodes (12): Basic Usage, Connection String Formats, Dry Run Mode, Environment Variables, Migration, Migration Command Line Options, SQLite Usage, SSH Tunnel Support (+4 more)

### Community 6 - "SQL Toolbox"
Cohesion: 0.20
Nodes (9): Example Directory Structure, Features, From Source, Installation, License, Migration History, Script Organization, SQL Toolbox (+1 more)

### Community 7 - "NewDatabaseDumper"
Cohesion: 0.12
Nodes (15): NewPgDumpGoDumper(), NewDatabaseDumper(), NewPgDumpBinaryDumper(), TestNewDatabaseDumper_GoMethodRequiresDB(), TestNewDatabaseDumper_InvalidMethod(), TestNewDatabaseDumper_PgDumpMethodCreatesInstance(), TestNewDatabaseDumper_PgDumpMethodRequiresConnParams(), TestPgDumpGoDumper_NilDB() (+7 more)

### Community 8 - "Multi-Database Backup (`backup` Subcommand)"
Cohesion: 0.25
Nodes (8): Backup Output, Backup Subcommand Flags, Differences from `--backup-only` Flag, Multi-Database Backup (`backup` Subcommand), Quick Start, Unified YAML Configuration, What Gets Dumped, Wildcard Database Backup

### Community 9 - "CLAUDE.md — sql-toolbox"
Cohesion: 0.29
Nodes (6): Architecture, Build / Test Commands, CLAUDE.md — sql-toolbox, Code Style, graphify, Key Features / Gotchas

### Community 10 - "DatabaseTarget"
Cohesion: 0.06
Nodes (29): ExpandWildcardDatabasesWithConnStr(), BackupConfig, DatabaseTarget, S3Config, QueryDatabases(), TestBackupConfig_ExpandWildcards_NoWildcards(), TestBackupConfig_Validate(), TestDatabaseTarget_ToConnectionString_OmittedPort() (+21 more)

### Community 11 - "integration-test.sh"
Cohesion: 0.40
Nodes (4): integration-test.sh script, SQLTB_DATABASE, SQLTB_PROVIDER, SQLTB_VERBOSE

### Community 12 - "SSHTunnel"
Cohesion: 0.14
Nodes (4): SSHTunnelConfig, EstablishIfNeeded(), NewSSHTunnel(), SSHTunnel

### Community 13 - "ConnectionParameters"
Cohesion: 0.07
Nodes (22): Factory, NewFactory(), convertDotNetFormat(), convertPostgreSQLURL(), ConnectionParameters, DatabaseProvider, normalizeConnectionString(), ParseDatabaseProvider() (+14 more)

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

## Knowledge Gaps
- **4 isolated node(s):** `github.com/adeotek/adeotek-tools/sql-toolbox`, `SQLTB_PROVIDER`, `SQLTB_DATABASE`, `SQLTB_VERBOSE`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 107 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `DatabaseTarget` connect `DatabaseTarget` to `migration_cmd.go`, `SSHTunnel`?**
  _High betweenness centrality (0.058) - this node is a cross-community bridge._
- **What connects `github.com/adeotek/adeotek-tools/sql-toolbox`, `SQLTB_PROVIDER`, `SQLTB_DATABASE` to the rest of the system?**
  _4 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.0624048706240487 - nodes in this community are weakly interconnected._
- **Why does `SSHTunnel` connect `SSHTunnel` to `migration_cmd.go`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **Should `runMigration` be split into smaller, more focused modules?**
  _Cohesion score 0.13043478260869565 - nodes in this community are weakly interconnected._
- **Why does `ConnectionParameters` connect `ConnectionParameters` to `migration_cmd.go`, `SSHTunnel`, `runMigration`, `NewDatabaseDumper`?**
  _High betweenness centrality (0.051) - this node is a cross-community bridge._
- **Should `NewDatabaseDumper` be split into smaller, more focused modules?**
  _Cohesion score 0.12 - nodes in this community are weakly interconnected._