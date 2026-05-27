# Pure Go pg_dump Impl

> 30 nodes

## Key Concepts

- **PgDumpService** (11 connections) — `internal/services/pg_dump_schema.go`
- **qualifiedName()** (11 connections) — `internal/services/pg_dump_service.go`
- **quoteIdentifier()** (9 connections) — `internal/services/pg_dump_service.go`
- **.dumpTableData()** (5 connections) — `internal/services/pg_dump_data.go`
- **formatInsertValue()** (5 connections) — `internal/services/pg_dump_data.go`
- **pg_dump_service.go** (5 connections) — `internal/services/pg_dump_service.go`
- **pg_dump_service_test.go** (5 connections) — `internal/services/pg_dump_service_test.go`
- **.dumpEnumTypes()** (4 connections) — `internal/services/pg_dump_schema.go`
- **.dumpSingleTable()** (4 connections) — `internal/services/pg_dump_schema.go`
- **escapeStringLiteral()** (3 connections) — `internal/services/pg_dump_data.go`
- **.dumpForeignKeys()** (3 connections) — `internal/services/pg_dump_schema.go`
- **.dumpTriggers()** (3 connections) — `internal/services/pg_dump_schema.go`
- **parsePostgresArray()** (3 connections) — `internal/services/pg_dump_schema.go`
- **pg_dump_data.go** (2 connections) — `internal/services/pg_dump_data.go`
- **PgDumpService** (2 connections) — `internal/services/pg_dump_data.go`
- **.dumpAllTableData()** (2 connections) — `internal/services/pg_dump_data.go`
- **.dumpExtensions()** (2 connections) — `internal/services/pg_dump_schema.go`
- **.dumpSchemas()** (2 connections) — `internal/services/pg_dump_schema.go`
- **.dumpSequences()** (2 connections) — `internal/services/pg_dump_schema.go`
- **.dumpTables()** (2 connections) — `internal/services/pg_dump_schema.go`
- **.dumpViews()** (2 connections) — `internal/services/pg_dump_schema.go`
- **.dumpIndexes()** (2 connections) — `internal/services/pg_dump_schema.go`
- **TestEscapeStringLiteral()** (2 connections) — `internal/services/pg_dump_service_test.go`
- **TestFormatInsertValue()** (2 connections) — `internal/services/pg_dump_service_test.go`
- **TestParsePostgresArray()** (2 connections) — `internal/services/pg_dump_service_test.go`
- *... and 5 more nodes in this community*

## Relationships

- [[S3 Upload & Backup Service]] (2 shared connections)
- [[pg_dump Service Logic]] (1 shared connections)
- [[Database Dumper Strategy]] (1 shared connections)

## Source Files

- `internal/services/pg_dump_data.go`
- `internal/services/pg_dump_schema.go`
- `internal/services/pg_dump_service.go`
- `internal/services/pg_dump_service_test.go`

## Audit Trail

- EXTRACTED: 62 (61%)
- INFERRED: 40 (39%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*