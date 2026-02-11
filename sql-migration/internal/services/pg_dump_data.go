package services

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

// dumpAllTableData dumps INSERT statements for all included tables
func (s *PgDumpService) dumpAllTableData() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'r'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND c.relname NOT LIKE 'pg_%'
		ORDER BY n.nspname, c.relname`)
	if err != nil {
		return fmt.Errorf("querying tables for data dump: %w", err)
	}
	defer rows.Close()

	type tableInfo struct {
		schema, name string
	}
	var tables []tableInfo
	for rows.Next() {
		var ti tableInfo
		if err := rows.Scan(&ti.schema, &ti.name); err != nil {
			return fmt.Errorf("scanning table for data dump: %w", err)
		}
		if !s.isSchemaIncluded(ti.schema) {
			continue
		}
		if s.isTableExcluded(ti.name) {
			continue
		}
		tables = append(tables, ti)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(tables) == 0 {
		return nil
	}

	if err := s.writeln("\n--\n-- Data\n--\n"); err != nil {
		return err
	}

	for _, tbl := range tables {
		if err := s.dumpTableData(tbl.schema, tbl.name); err != nil {
			return fmt.Errorf("dumping data for %s.%s: %w", tbl.schema, tbl.name, err)
		}
	}
	return nil
}

// dumpTableData dumps INSERT statements for a single table
func (s *PgDumpService) dumpTableData(schema, table string) error {
	qname := qualifiedName(schema, table)

	// Get column names and types
	colRows, err := s.db.Query(`
		SELECT a.attname, format_type(a.atttypid, a.atttypmod) AS data_type
		FROM pg_attribute a
		WHERE a.attrelid = $1::regclass
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY a.attnum`, qname)
	if err != nil {
		return fmt.Errorf("querying columns: %w", err)
	}
	defer colRows.Close()

	var colNames []string
	var colTypes []string
	for colRows.Next() {
		var name, dtype string
		if err := colRows.Scan(&name, &dtype); err != nil {
			return fmt.Errorf("scanning column info: %w", err)
		}
		colNames = append(colNames, name)
		colTypes = append(colTypes, dtype)
	}
	if err := colRows.Err(); err != nil {
		return err
	}

	if len(colNames) == 0 {
		return nil
	}

	// Select all data
	quotedCols := make([]string, len(colNames))
	for i, c := range colNames {
		quotedCols[i] = quoteIdentifier(c)
	}
	selectSQL := fmt.Sprintf("SELECT %s FROM %s", strings.Join(quotedCols, ", "), qname)

	dataRows, err := s.db.Query(selectSQL)
	if err != nil {
		return fmt.Errorf("querying data: %w", err)
	}
	defer dataRows.Close()

	insertPrefix := fmt.Sprintf("INSERT INTO %s (%s) VALUES", qname, strings.Join(quotedCols, ", "))

	hasData := false
	for dataRows.Next() {
		if !hasData {
			if err := s.writef("--\n-- Data for table %s\n--\n\n", qname); err != nil {
				return err
			}
			hasData = true
		}

		// Scan into interface{} slice
		values := make([]interface{}, len(colNames))
		valuePtrs := make([]interface{}, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := dataRows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}

		// Format values
		formattedValues := make([]string, len(values))
		for i, val := range values {
			formattedValues[i] = formatInsertValue(val, colTypes[i])
		}

		if err := s.writef("%s (%s);\n", insertPrefix, strings.Join(formattedValues, ", ")); err != nil {
			return err
		}
	}
	if hasData {
		if err := s.writeln(""); err != nil {
			return err
		}
	}
	return dataRows.Err()
}

// formatInsertValue formats a Go value for inclusion in a PostgreSQL INSERT statement
func formatInsertValue(value interface{}, colType string) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if math.IsNaN(v) {
			return "'NaN'"
		}
		if math.IsInf(v, 1) {
			return "'Infinity'"
		}
		if math.IsInf(v, -1) {
			return "'-Infinity'"
		}
		return fmt.Sprintf("%g", v)
	case []byte:
		// Check if this is a bytea column
		if strings.HasPrefix(colType, "bytea") {
			return fmt.Sprintf("'\\x%s'", hex.EncodeToString(v))
		}
		// Otherwise treat as string
		return "'" + escapeStringLiteral(string(v)) + "'"
	case string:
		return "'" + escapeStringLiteral(v) + "'"
	case time.Time:
		return "'" + v.Format("2006-01-02 15:04:05.999999-07") + "'"
	default:
		// Fallback: convert to string
		return "'" + escapeStringLiteral(fmt.Sprintf("%v", v)) + "'"
	}
}

// escapeStringLiteral escapes a string for use in a PostgreSQL string literal
func escapeStringLiteral(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}
