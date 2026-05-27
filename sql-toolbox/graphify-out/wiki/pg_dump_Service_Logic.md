# pg_dump Service Logic

> 8 nodes

## Key Concepts

- **PgDumpService** (8 connections) — `internal/services/pg_dump_service.go`
- **.Dump()** (3 connections) — `internal/services/pg_dump_service.go`
- **.writeHeader()** (2 connections) — `internal/services/pg_dump_service.go`
- **.writeFooter()** (2 connections) — `internal/services/pg_dump_service.go`
- **.writef()** (1 connections) — `internal/services/pg_dump_service.go`
- **.writeln()** (1 connections) — `internal/services/pg_dump_service.go`
- **.isSchemaIncluded()** (1 connections) — `internal/services/pg_dump_service.go`
- **.isTableExcluded()** (1 connections) — `internal/services/pg_dump_service.go`

## Relationships

- [[Pure Go pg_dump Impl]] (1 shared connections)

## Source Files

- `internal/services/pg_dump_service.go`

## Audit Trail

- EXTRACTED: 19 (100%)
- INFERRED: 0 (0%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*