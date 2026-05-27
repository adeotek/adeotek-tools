# CLI App Commands

> 21 nodes

## Key Concepts

- **runMigration()** (10 connections) — `internal/app/migration_cmd.go`
- **migration_cmd.go** (5 connections) — `internal/app/migration_cmd.go`
- **EstablishIfNeeded()** (5 connections) — `internal/tunnel/helper.go`
- **runBackup()** (4 connections) — `internal/app/backup_cmd.go`
- **ExpandWildcardDatabasesWithTunnel()** (4 connections) — `internal/services/wildcard_helper.go`
- **context.go** (3 connections) — `internal/app/context.go`
- **newMigrationCmd()** (3 connections) — `internal/app/migration_cmd.go`
- **Run()** (3 connections) — `internal/app/app.go`
- **ExpandWildcardsWithTunnels()** (3 connections) — `internal/services/wildcard_helper.go`
- **backup_cmd.go** (2 connections) — `internal/app/backup_cmd.go`
- **newBackupCmd()** (2 connections) — `internal/app/backup_cmd.go`
- **setViper()** (2 connections) — `internal/app/context.go`
- **getViper()** (2 connections) — `internal/app/context.go`
- **resolveString()** (2 connections) — `internal/app/migration_cmd.go`
- **resolveInt()** (2 connections) — `internal/app/migration_cmd.go`
- **resolveBool()** (2 connections) — `internal/app/migration_cmd.go`
- **ParseConnectionParameters()** (2 connections) — `internal/models/connection.go`
- **wildcard_helper.go** (2 connections) — `internal/services/wildcard_helper.go`
- **viperKey** (1 connections) — `internal/app/context.go`
- **app.go** (1 connections) — `internal/app/app.go`
- **helper.go** (1 connections) — `internal/tunnel/helper.go`

## Relationships

- [[Config Loader]] (2 shared connections)
- [[Backup Orchestration]] (2 shared connections)
- [[S3 Upload & Backup Service]] (1 shared connections)
- [[Migration Engine]] (1 shared connections)
- [[Connection String Handling]] (1 shared connections)
- [[Backup Config Models]] (1 shared connections)
- [[SSH Tunnel]] (1 shared connections)

## Source Files

- `internal/app/app.go`
- `internal/app/backup_cmd.go`
- `internal/app/context.go`
- `internal/app/migration_cmd.go`
- `internal/models/connection.go`
- `internal/services/wildcard_helper.go`
- `internal/tunnel/helper.go`

## Audit Trail

- EXTRACTED: 37 (61%)
- INFERRED: 24 (39%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*