# SQL Script Scanning & Hashing

> 19 nodes

## Key Concepts

- **NewSqlScriptsHelpers()** (11 connections) — `internal/services/sql_scripts_helpers.go`
- **sql_scripts_helpers_test.go** (10 connections) — `internal/services/sql_scripts_helpers_test.go`
- **SqlScriptsHelpers** (6 connections) — `internal/services/sql_scripts_helpers.go`
- **.ExecuteScript()** (3 connections) — `internal/services/sql_scripts_helpers.go`
- **sql_scripts_helpers.go** (2 connections) — `internal/services/sql_scripts_helpers.go`
- **.ScanForSqlFiles()** (2 connections) — `internal/services/sql_scripts_helpers.go`
- **.scanDirectoryBreadthFirst()** (2 connections) — `internal/services/sql_scripts_helpers.go`
- **.SplitSqlStatements()** (2 connections) — `internal/services/sql_scripts_helpers.go`
- **TestSqlScriptsHelpers_ScanForSqlFiles()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_CalculateHash()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_ScanForSqlFiles_Comprehensive()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_ScanForSqlFiles_NestedDirectories()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_ScanForSqlFiles_ErrorCases()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_SplitSqlStatements_ComplexScript()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_SplitSqlStatements_SimpleStatements()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_SplitSqlStatements_WithComments()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **TestSqlScriptsHelpers_SplitSqlStatements_DollarEdgeCases()** (2 connections) — `internal/services/sql_scripts_helpers_test.go`
- **.CalculateHash()** (1 connections) — `internal/services/sql_scripts_helpers.go`
- **TestSqlScriptsHelpers_DollarQuoteRegexPattern()** (1 connections) — `internal/services/sql_scripts_helpers_test.go`

## Relationships

- [[Migration Engine]] (2 shared connections)

## Source Files

- `internal/services/sql_scripts_helpers.go`
- `internal/services/sql_scripts_helpers_test.go`

## Audit Trail

- EXTRACTED: 38 (66%)
- INFERRED: 20 (34%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*