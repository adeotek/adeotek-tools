# Graph Report - .  (2026-05-26)

## Corpus Check
- Corpus is ~26,341 words - fits in a single context window. You may not need a graph.

## Summary
- 308 nodes · 468 edges · 19 communities (13 shown, 6 thin omitted)
- Extraction: 75% EXTRACTED · 25% INFERRED · 0% AMBIGUOUS · INFERRED: 117 edges (avg confidence: 0.81)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Build, Test & CLI Docs|Build, Test & CLI Docs]]
- [[_COMMUNITY_Backup Config Models|Backup Config Models]]
- [[_COMMUNITY_Pure Go pg_dump Impl|Pure Go pg_dump Impl]]
- [[_COMMUNITY_S3 Upload & Backup Service|S3 Upload & Backup Service]]
- [[_COMMUNITY_CLI App Commands|CLI App Commands]]
- [[_COMMUNITY_Connection String Handling|Connection String Handling]]
- [[_COMMUNITY_SQL Script Scanning & Hashing|SQL Script Scanning & Hashing]]
- [[_COMMUNITY_Database Dumper Strategy|Database Dumper Strategy]]
- [[_COMMUNITY_Backup Orchestration|Backup Orchestration]]
- [[_COMMUNITY_Migration Engine|Migration Engine]]
- [[_COMMUNITY_Database Target Model|Database Target Model]]
- [[_COMMUNITY_Config Loader|Config Loader]]
- [[_COMMUNITY_SSH Tunnel|SSH Tunnel]]
- [[_COMMUNITY_Backup Service Core|Backup Service Core]]
- [[_COMMUNITY_pg_dump Service Logic|pg_dump Service Logic]]
- [[_COMMUNITY_SSH Tunnel Config|SSH Tunnel Config]]
- [[_COMMUNITY_Script Execution History|Script Execution History]]

## God Nodes (most connected - your core abstractions)
1. `DatabaseTarget` - 16 edges
2. `backup subcommand` - 15 edges
3. `writeTempConfig()` - 14 edges
4. `migration subcommand` - 14 edges
5. `SQL Toolbox README` - 13 edges
6. `LoadConfig()` - 11 edges
7. `PgDumpService` - 11 edges
8. `qualifiedName()` - 11 edges
9. `NewSqlScriptsHelpers()` - 11 edges
10. `SSHTunnel` - 11 edges

## Surprising Connections (you probably didn't know these)
- `migration subcommand` --references--> `Dollar-Quoting SQL Parser ($$, $body$, $inner$, $func$)`  [INFERRED]
  sql-toolbox/README.md → sql-toolbox/test-scripts/stored_procedures/001_complex_procedure.sql
- `Ordered Script Execution (alphabetical + recursive subdirs)` --conceptually_related_to--> `001_deeply_nested.sql - Deeply nested directory test script`  [INFERRED]
  sql-toolbox/README.md → sql-toolbox/test-scripts/tables/level1/level2/level3/001_deeply_nested.sql
- `runBackup()` --calls--> `LoadConfig()`  [INFERRED]
  internal/app/backup_cmd.go → internal/models/config.go
- `runBackup()` --calls--> `NewBackupOnlyOrchestrator()`  [INFERRED]
  internal/app/backup_cmd.go → internal/services/backup_only_orchestrator.go
- `runMigration()` --calls--> `getViper()`  [INFERRED]
  internal/app/migration_cmd.go → internal/app/context.go

## Hyperedges (group relationships)
- **Test Script Execution Order (tables → views → stored_procedures → data)** — test_scripts_tables_001_create_users, test_scripts_tables_002_create_posts, test_scripts_views_001_user_posts_view, test_scripts_stored_procedures_001_complex_procedure, test_scripts_data_001_seed_users, test_scripts_data_002_seed_posts [EXTRACTED 1.00]
- **Dual Backup Methods: Pure Go vs pg_dump** — sql_toolbox_pure_go_dump, sql_toolbox_pg_dump_binary, sql_toolbox_backup_subcommand [EXTRACTED 1.00]
- **S3-Compatible Storage Backends (AWS S3, MinIO, RustFS)** — sql_toolbox_s3_upload, sql_toolbox_s3_per_db_overrides, sql_toolbox_backup_subcommand [EXTRACTED 1.00]
- **Dollar-Quoting Edge Case Test Coverage** — test_scripts_stored_procedures_001_complex_procedure, test_scripts_complex_dollar_test, sql_toolbox_dollar_quoting [EXTRACTED 1.00]

## Communities (19 total, 6 thin omitted)

### Community 0 - "Build, Test & CLI Docs"
Cohesion: 0.09
Nodes (35): backup_worker.sh script, build.sh script, integration-test.sh script, 001_deeply_nested.sql - Deeply nested directory test script, backup subcommand, SQL Toolbox CLAUDE.md, Connection String Formats (lib/pq, .NET, PostgreSQL URL), Dollar-Quoting SQL Parser ($$, $body$, $inner$, $func$) (+27 more)

### Community 1 - "Backup Config Models"
Cohesion: 0.09
Nodes (23): ExpandWildcardDatabasesWithConnStr(), LoadBackupConfig(), QueryDatabases(), boolPtr(), containsAny(), strPtr(), TestDatabaseTarget_GetEffectiveBackupMethod(), TestDatabaseTarget_GetEffectiveCompress() (+15 more)

### Community 2 - "Pure Go pg_dump Impl"
Cohesion: 0.11
Nodes (13): PgDumpService, PgDumpService, escapeStringLiteral(), formatInsertValue(), parsePostgresArray(), qualifiedName(), quoteIdentifier(), TestEscapeStringLiteral() (+5 more)

### Community 3 - "S3 Upload & Backup Service"
Cohesion: 0.13
Nodes (16): NewBackupService(), TestBackupService_CreateBackup_DryRun(), TestBackupService_CreateBackup_SQLite(), TestBackupService_GetDatabaseName(), TestBackupService_GetLastBackupPath(), TestBackupService_GetLastBackupPath_NoBackup(), TestBackupService_RestoreLastBackup_NoBackup(), TestBackupService_RestoreLastBackup_SQLite() (+8 more)

### Community 4 - "CLI App Commands"
Cohesion: 0.12
Nodes (15): Run(), newBackupCmd(), runBackup(), getViper(), setViper(), newMigrationCmd(), resolveBool(), resolveInt() (+7 more)

### Community 5 - "Connection String Handling"
Cohesion: 0.15
Nodes (10): convertDotNetFormat(), convertPostgreSQLURL(), normalizeConnectionString(), ParseDatabaseProvider(), TestNormalizeConnectionString_DotNetFormat(), TestNormalizeConnectionString_LibPQFormat(), TestNormalizeConnectionString_PostgreSQLURL(), TestParseDatabaseProvider() (+2 more)

### Community 6 - "SQL Script Scanning & Hashing"
Cohesion: 0.16
Nodes (11): NewSqlScriptsHelpers(), TestSqlScriptsHelpers_CalculateHash(), TestSqlScriptsHelpers_ScanForSqlFiles(), TestSqlScriptsHelpers_ScanForSqlFiles_Comprehensive(), TestSqlScriptsHelpers_ScanForSqlFiles_ErrorCases(), TestSqlScriptsHelpers_ScanForSqlFiles_NestedDirectories(), TestSqlScriptsHelpers_SplitSqlStatements_ComplexScript(), TestSqlScriptsHelpers_SplitSqlStatements_DollarEdgeCases() (+3 more)

### Community 7 - "Database Dumper Strategy"
Cohesion: 0.13
Nodes (11): DatabaseDumper, NewPgDumpGoDumper(), NewDatabaseDumper(), NewPgDumpBinaryDumper(), TestNewDatabaseDumper_GoMethodRequiresDB(), TestNewDatabaseDumper_InvalidMethod(), TestNewDatabaseDumper_PgDumpMethodCreatesInstance(), TestNewDatabaseDumper_PgDumpMethodRequiresConnParams() (+3 more)

### Community 8 - "Backup Orchestration"
Cohesion: 0.21
Nodes (13): createOutputWriter(), NewBackupOnlyOrchestrator(), boolP(), linkConfigDefaults(), TestBackupOnlyOrchestrator_ConnectionFailure(), TestBackupOnlyOrchestrator_DryRun(), TestBackupOnlyOrchestrator_DryRunWithS3(), TestBackupOnlyOrchestrator_PartialFailure() (+5 more)

### Community 9 - "Migration Engine"
Cohesion: 0.12
Nodes (6): Factory, NewFactory(), NewMigrationHistoryRepository(), MigrationHistoryRepository, NewMigrationService(), MigrationService

### Community 11 - "Config Loader"
Cohesion: 0.24
Nodes (11): Config, LoadConfig(), TestLoadConfig_BackupOnly(), TestLoadConfig_EmptyFile(), TestLoadConfig_FileNotFound(), TestLoadConfig_InvalidYAML(), TestLoadConfig_MigrationBackupMethod(), TestLoadConfig_MigrationDefaults() (+3 more)

## Knowledge Gaps
- **24 isolated node(s):** `backup_worker.sh script`, `build.sh script`, `integration-test.sh script`, `SQLTB_PROVIDER`, `SQLTB_DATABASE` (+19 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `runMigration()` connect `CLI App Commands` to `S3 Upload & Backup Service`, `Migration Engine`, `Config Loader`?**
  _High betweenness centrality (0.253) - this node is a cross-community bridge._
- **Why does `EstablishIfNeeded()` connect `CLI App Commands` to `Backup Orchestration`, `SSH Tunnel`?**
  _High betweenness centrality (0.176) - this node is a cross-community bridge._
- **Why does `LoadConfig()` connect `Config Loader` to `CLI App Commands`?**
  _High betweenness centrality (0.139) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `writeTempConfig()` (e.g. with `TestLoadConfig_MigrationOnly()` and `TestLoadConfig_BackupOnly()`) actually correct?**
  _`writeTempConfig()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `migration subcommand` (e.g. with `build.sh` and `Dollar-Quoting SQL Parser ($$, $body$, $inner$, $func$)`) actually correct?**
  _`migration subcommand` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `backup_worker.sh script`, `build.sh script`, `integration-test.sh script` to the rest of the system?**
  _24 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build, Test & CLI Docs` be split into smaller, more focused modules?**
  _Cohesion score 0.09103840682788052 - nodes in this community are weakly interconnected._