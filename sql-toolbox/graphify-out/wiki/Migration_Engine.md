# Migration Engine

> 16 nodes

## Key Concepts

- **MigrationHistoryRepository** (5 connections) — `internal/repository/migration_history.go`
- **.Run()** (4 connections) — `internal/services/migration_service.go`
- **Factory** (3 connections) — `internal/database/factory.go`
- **NewFactory()** (3 connections) — `internal/database/factory.go`
- **NewMigrationService()** (3 connections) — `internal/services/migration_service.go`
- **factory.go** (2 connections) — `internal/database/factory.go`
- **migration_history.go** (2 connections) — `internal/repository/migration_history.go`
- **NewMigrationHistoryRepository()** (2 connections) — `internal/repository/migration_history.go`
- **migration_service.go** (2 connections) — `internal/services/migration_service.go`
- **MigrationService** (2 connections) — `internal/services/migration_service.go`
- **.CreateConnection()** (1 connections) — `internal/database/factory.go`
- **.GetProvider()** (1 connections) — `internal/database/factory.go`
- **.IsHistoryTableCreated()** (1 connections) — `internal/repository/migration_history.go`
- **.CreateHistoryTable()** (1 connections) — `internal/repository/migration_history.go`
- **.GetExecutedScripts()** (1 connections) — `internal/repository/migration_history.go`
- **.UpsertExecutedScript()** (1 connections) — `internal/repository/migration_history.go`

## Relationships

- [[SQL Script Scanning & Hashing]] (2 shared connections)
- [[CLI App Commands]] (1 shared connections)
- [[S3 Upload & Backup Service]] (1 shared connections)

## Source Files

- `internal/database/factory.go`
- `internal/repository/migration_history.go`
- `internal/services/migration_service.go`

## Audit Trail

- EXTRACTED: 26 (76%)
- INFERRED: 8 (24%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*