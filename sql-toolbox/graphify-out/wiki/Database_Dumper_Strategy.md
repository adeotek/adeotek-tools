# Database Dumper Strategy

> 18 nodes

## Key Concepts

- **NewDatabaseDumper()** (9 connections) — `internal/services/dumper.go`
- **dumper_test.go** (5 connections) — `internal/services/dumper_test.go`
- **dumper.go** (2 connections) — `internal/services/dumper.go`
- **dumper_go.go** (2 connections) — `internal/services/dumper_go.go`
- **PgDumpGoDumper** (2 connections) — `internal/services/dumper_go.go`
- **NewPgDumpGoDumper()** (2 connections) — `internal/services/dumper_go.go`
- **.Dump()** (2 connections) — `internal/services/dumper_go.go`
- **dumper_pgdump.go** (2 connections) — `internal/services/dumper_pgdump.go`
- **PgDumpBinaryDumper** (2 connections) — `internal/services/dumper_pgdump.go`
- **NewPgDumpBinaryDumper()** (2 connections) — `internal/services/dumper_pgdump.go`
- **TestNewDatabaseDumper_InvalidMethod()** (2 connections) — `internal/services/dumper_test.go`
- **TestNewDatabaseDumper_GoMethodRequiresDB()** (2 connections) — `internal/services/dumper_test.go`
- **TestNewDatabaseDumper_PgDumpMethodRequiresConnParams()** (2 connections) — `internal/services/dumper_test.go`
- **TestNewDatabaseDumper_PgDumpMethodCreatesInstance()** (2 connections) — `internal/services/dumper_test.go`
- **NewPgDumpService()** (2 connections) — `internal/services/pg_dump_service.go`
- **DatabaseDumper** (1 connections) — `internal/services/dumper.go`
- **.Dump()** (1 connections) — `internal/services/dumper_pgdump.go`
- **TestPgDumpGoDumper_NilDB()** (1 connections) — `internal/services/dumper_test.go`

## Relationships

- [[Backup Service Core]] (1 shared connections)
- [[Backup Orchestration]] (1 shared connections)
- [[Pure Go pg_dump Impl]] (1 shared connections)

## Source Files

- `internal/services/dumper.go`
- `internal/services/dumper_go.go`
- `internal/services/dumper_pgdump.go`
- `internal/services/dumper_test.go`
- `internal/services/pg_dump_service.go`

## Audit Trail

- EXTRACTED: 27 (63%)
- INFERRED: 16 (37%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*