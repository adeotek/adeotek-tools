# Backup Config Models

> 37 nodes

## Key Concepts

- **backup_config_test.go** (25 connections) — `internal/models/backup_config_test.go`
- **writeTempConfig()** (14 connections) — `internal/models/backup_config_test.go`
- **LoadBackupConfig()** (10 connections) — `internal/models/backup_config.go`
- **backup_config.go** (7 connections) — `internal/models/backup_config.go`
- **ExpandWildcardDatabasesWithConnStr()** (6 connections) — `internal/models/backup_config.go`
- **boolPtr()** (4 connections) — `internal/models/backup_config_test.go`
- **BackupConfig** (3 connections) — `internal/models/backup_config.go`
- **.SetDefaults()** (3 connections) — `internal/models/backup_config.go`
- **.ToConnectionString()** (3 connections) — `internal/models/backup_config.go`
- **QueryDatabases()** (3 connections) — `internal/models/backup_config.go`
- **.ExpandWildcards()** (3 connections) — `internal/models/backup_config.go`
- **TestLoadBackupConfig_ValidConfig()** (3 connections) — `internal/models/backup_config_test.go`
- **TestLoadBackupConfig_NoDatabases()** (3 connections) — `internal/models/backup_config_test.go`
- **TestLoadBackupConfig_MissingName()** (3 connections) — `internal/models/backup_config_test.go`
- **TestLoadBackupConfig_MissingHostAndConnectionString()** (3 connections) — `internal/models/backup_config_test.go`
- **TestLoadBackupConfig_ConnectionString()** (3 connections) — `internal/models/backup_config_test.go`
- **TestLoadBackupConfig_S3EnabledNoBucket()** (3 connections) — `internal/models/backup_config_test.go`
- **TestExpandWildcardDatabases_EmptyExcludeDb()** (3 connections) — `internal/models/backup_config_test.go`
- **TestLoadBackupConfig_WildcardDatabase()** (3 connections) — `internal/models/backup_config_test.go`
- **.validate()** (2 connections) — `internal/models/backup_config.go`
- **TestDatabaseTarget_GetEffectiveCompress()** (2 connections) — `internal/models/backup_config_test.go`
- **TestDatabaseTarget_GetEffectiveUploadToS3()** (2 connections) — `internal/models/backup_config_test.go`
- **TestDatabaseTarget_GetEffectiveDeleteLocalAfterUpload()** (2 connections) — `internal/models/backup_config_test.go`
- **strPtr()** (2 connections) — `internal/models/backup_config_test.go`
- **TestDatabaseTarget_GetEffectiveBackupMethod()** (2 connections) — `internal/models/backup_config_test.go`
- *... and 12 more nodes in this community*

## Relationships

- [[Config Loader]] (6 shared connections)
- [[Database Target Model]] (3 shared connections)
- [[Connection String Handling]] (1 shared connections)
- [[CLI App Commands]] (1 shared connections)

## Source Files

- `internal/models/backup_config.go`
- `internal/models/backup_config_test.go`

## Audit Trail

- EXTRACTED: 105 (80%)
- INFERRED: 26 (20%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*