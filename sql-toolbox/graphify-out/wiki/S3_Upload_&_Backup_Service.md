# S3 Upload & Backup Service

> 22 nodes

## Key Concepts

- **.String()** (10 connections) — `internal/models/connection.go`
- **NewBackupService()** (9 connections) — `internal/services/backup_service.go`
- **backup_service_test.go** (7 connections) — `internal/services/backup_service_test.go`
- **s3_upload_service.go** (4 connections) — `internal/services/s3_upload_service.go`
- **s3_upload_service_test.go** (4 connections) — `internal/services/s3_upload_service_test.go`
- **TestBackupService_CreateBackup_SQLite()** (3 connections) — `internal/services/backup_service_test.go`
- **TestBackupService_RestoreLastBackup_SQLite()** (3 connections) — `internal/services/backup_service_test.go`
- **NewS3UploadService()** (3 connections) — `internal/services/s3_upload_service.go`
- **BuildS3Key()** (3 connections) — `internal/services/s3_upload_service.go`
- **TestBackupService_CreateBackup_DryRun()** (2 connections) — `internal/services/backup_service_test.go`
- **TestBackupService_GetLastBackupPath()** (2 connections) — `internal/services/backup_service_test.go`
- **TestBackupService_GetLastBackupPath_NoBackup()** (2 connections) — `internal/services/backup_service_test.go`
- **TestBackupService_RestoreLastBackup_NoBackup()** (2 connections) — `internal/services/backup_service_test.go`
- **TestBackupService_GetDatabaseName()** (2 connections) — `internal/services/backup_service_test.go`
- **S3UploadService** (2 connections) — `internal/services/s3_upload_service.go`
- **.Upload()** (2 connections) — `internal/services/s3_upload_service.go`
- **MockS3Uploader** (2 connections) — `internal/services/s3_upload_service_test.go`
- **.Upload()** (2 connections) — `internal/services/s3_upload_service_test.go`
- **TestBuildS3Key()** (2 connections) — `internal/services/s3_upload_service_test.go`
- **TestMockS3Uploader()** (2 connections) — `internal/services/s3_upload_service_test.go`
- **S3Uploader** (1 connections) — `internal/services/s3_upload_service.go`
- **mockUploadCall** (1 connections) — `internal/services/s3_upload_service_test.go`

## Relationships

- [[Backup Orchestration]] (3 shared connections)
- [[Backup Service Core]] (2 shared connections)
- [[Pure Go pg_dump Impl]] (2 shared connections)
- [[Connection String Handling]] (1 shared connections)
- [[Migration Engine]] (1 shared connections)
- [[CLI App Commands]] (1 shared connections)

## Source Files

- `internal/models/connection.go`
- `internal/services/backup_service.go`
- `internal/services/backup_service_test.go`
- `internal/services/s3_upload_service.go`
- `internal/services/s3_upload_service_test.go`

## Audit Trail

- EXTRACTED: 38 (54%)
- INFERRED: 32 (46%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*