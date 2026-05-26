# Connection String Handling

> 20 nodes

## Key Concepts

- **normalizeConnectionString()** (8 connections) — `internal/models/connection.go`
- **connection.go** (7 connections) — `internal/models/connection.go`
- **connection_test.go** (7 connections) — `internal/models/connection_test.go`
- **ConnectionParameters** (6 connections) — `internal/models/connection.go`
- **.GetDbProvider()** (4 connections) — `internal/models/connection.go`
- **ParseDatabaseProvider()** (3 connections) — `internal/models/connection.go`
- **.GetConnectionString()** (3 connections) — `internal/models/connection.go`
- **DatabaseProvider** (2 connections) — `internal/models/connection.go`
- **.IsValid()** (2 connections) — `internal/models/connection.go`
- **convertPostgreSQLURL()** (2 connections) — `internal/models/connection.go`
- **convertDotNetFormat()** (2 connections) — `internal/models/connection.go`
- **TestParseDatabaseProvider()** (2 connections) — `internal/models/connection_test.go`
- **TestNormalizeConnectionString_LibPQFormat()** (2 connections) — `internal/models/connection_test.go`
- **TestNormalizeConnectionString_DotNetFormat()** (2 connections) — `internal/models/connection_test.go`
- **TestNormalizeConnectionString_PostgreSQLURL()** (2 connections) — `internal/models/connection_test.go`
- **.SetActiveTunnel()** (1 connections) — `internal/models/connection.go`
- **.GetActiveTunnel()** (1 connections) — `internal/models/connection.go`
- **TestConnectionParameters_IsValid()** (1 connections) — `internal/models/connection_test.go`
- **TestConnectionParameters_GetConnectionString()** (1 connections) — `internal/models/connection_test.go`
- **TestGetConnectionString_WithNormalization()** (1 connections) — `internal/models/connection_test.go`

## Relationships

- [[CLI App Commands]] (1 shared connections)
- [[S3 Upload & Backup Service]] (1 shared connections)
- [[Backup Config Models]] (1 shared connections)

## Source Files

- `internal/models/connection.go`
- `internal/models/connection_test.go`

## Audit Trail

- EXTRACTED: 50 (85%)
- INFERRED: 9 (15%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*