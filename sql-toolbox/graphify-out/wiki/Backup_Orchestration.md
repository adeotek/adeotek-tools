# Backup Orchestration

> 17 nodes

## Key Concepts

- **backup_only_orchestrator_test.go** (8 connections) — `internal/services/backup_only_orchestrator_test.go`
- **NewBackupOnlyOrchestrator()** (7 connections) — `internal/services/backup_only_orchestrator.go`
- **.backupDatabase()** (6 connections) — `internal/services/backup_only_orchestrator.go`
- **linkConfigDefaults()** (5 connections) — `internal/services/backup_only_orchestrator_test.go`
- **backup_only_orchestrator.go** (5 connections) — `internal/services/backup_only_orchestrator.go`
- **TestBackupOnlyOrchestrator_DryRunWithS3()** (4 connections) — `internal/services/backup_only_orchestrator_test.go`
- **createOutputWriter()** (4 connections) — `internal/services/backup_only_orchestrator.go`
- **TestBackupOnlyOrchestrator_DryRun()** (3 connections) — `internal/services/backup_only_orchestrator_test.go`
- **TestBackupOnlyOrchestrator_ConnectionFailure()** (3 connections) — `internal/services/backup_only_orchestrator_test.go`
- **TestBackupOnlyOrchestrator_PartialFailure()** (3 connections) — `internal/services/backup_only_orchestrator_test.go`
- **TestCreateOutputWriter_Plain()** (3 connections) — `internal/services/backup_only_orchestrator_test.go`
- **BackupOnlyOrchestrator** (3 connections) — `internal/services/backup_only_orchestrator.go`
- **TestCreateOutputWriter_Gzip()** (2 connections) — `internal/services/backup_only_orchestrator_test.go`
- **boolP()** (2 connections) — `internal/services/backup_only_orchestrator_test.go`
- **.Run()** (2 connections) — `internal/services/backup_only_orchestrator.go`
- **DatabaseBackupResult** (1 connections) — `internal/services/backup_only_orchestrator.go`
- **BackupSummary** (1 connections) — `internal/services/backup_only_orchestrator.go`

## Relationships

- [[S3 Upload & Backup Service]] (3 shared connections)
- [[CLI App Commands]] (2 shared connections)
- [[Database Dumper Strategy]] (1 shared connections)

## Source Files

- `internal/services/backup_only_orchestrator.go`
- `internal/services/backup_only_orchestrator_test.go`

## Audit Trail

- EXTRACTED: 44 (71%)
- INFERRED: 18 (29%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*