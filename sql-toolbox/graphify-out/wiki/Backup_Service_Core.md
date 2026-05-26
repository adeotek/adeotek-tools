# Backup Service Core

> 11 nodes

## Key Concepts

- **BackupService** (10 connections) — `internal/services/backup_service.go`
- **.CreateBackup()** (5 connections) — `internal/services/backup_service.go`
- **.RestoreLastBackup()** (4 connections) — `internal/services/backup_service.go`
- **.GetLastBackupPath()** (3 connections) — `internal/services/backup_service.go`
- **.createPostgresBackup()** (3 connections) — `internal/services/backup_service.go`
- **.restorePostgresBackup()** (3 connections) — `internal/services/backup_service.go`
- **.getBackupDirectory()** (3 connections) — `internal/services/backup_service.go`
- **backup_service.go** (2 connections) — `internal/services/backup_service.go`
- **.createSQLiteBackup()** (2 connections) — `internal/services/backup_service.go`
- **.restoreSQLiteBackup()** (2 connections) — `internal/services/backup_service.go`
- **.getDatabaseName()** (2 connections) — `internal/services/backup_service.go`

## Relationships

- [[S3 Upload & Backup Service]] (2 shared connections)
- [[Database Dumper Strategy]] (1 shared connections)

## Source Files

- `internal/services/backup_service.go`

## Audit Trail

- EXTRACTED: 37 (95%)
- INFERRED: 2 (5%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*