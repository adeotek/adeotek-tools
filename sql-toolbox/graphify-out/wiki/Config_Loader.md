# Config Loader

> 13 nodes

## Key Concepts

- **LoadConfig()** (11 connections) — `internal/models/config.go`
- **config_test.go** (8 connections) — `internal/models/config_test.go`
- **config.go** (3 connections) — `internal/models/config.go`
- **TestLoadConfig_MigrationOnly()** (3 connections) — `internal/models/config_test.go`
- **TestLoadConfig_BackupOnly()** (3 connections) — `internal/models/config_test.go`
- **TestLoadConfig_UnifiedConfig()** (3 connections) — `internal/models/config_test.go`
- **TestLoadConfig_EmptyFile()** (3 connections) — `internal/models/config_test.go`
- **TestLoadConfig_InvalidYAML()** (3 connections) — `internal/models/config_test.go`
- **TestLoadConfig_MigrationDefaults()** (3 connections) — `internal/models/config_test.go`
- **TestLoadConfig_FileNotFound()** (2 connections) — `internal/models/config_test.go`
- **TestLoadConfig_MigrationBackupMethod()** (2 connections) — `internal/models/config_test.go`
- **Config** (1 connections) — `internal/models/config.go`
- **MigrationConfig** (1 connections) — `internal/models/config.go`

## Relationships

- [[Backup Config Models]] (6 shared connections)
- [[CLI App Commands]] (2 shared connections)

## Source Files

- `internal/models/config.go`
- `internal/models/config_test.go`

## Audit Trail

- EXTRACTED: 22 (48%)
- INFERRED: 24 (52%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*