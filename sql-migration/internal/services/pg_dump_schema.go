package services

import (
	"database/sql"
	"fmt"
	"strings"
)

// dumpExtensions dumps CREATE EXTENSION statements
func (s *PgDumpService) dumpExtensions() error {
	rows, err := s.db.Query(`SELECT extname FROM pg_extension WHERE extname != 'plpgsql' ORDER BY extname`)
	if err != nil {
		return fmt.Errorf("querying extensions: %w", err)
	}
	defer rows.Close()

	var hasExtensions bool
	for rows.Next() {
		if !hasExtensions {
			if err := s.writeln("\n--\n-- Extensions\n--\n"); err != nil {
				return err
			}
			hasExtensions = true
		}
		var extname string
		if err := rows.Scan(&extname); err != nil {
			return fmt.Errorf("scanning extension: %w", err)
		}
		if s.options.Clean {
			if err := s.writef("DROP EXTENSION IF EXISTS %s CASCADE;\n", quoteIdentifier(extname)); err != nil {
				return err
			}
		}
		if err := s.writef("CREATE EXTENSION IF NOT EXISTS %s WITH SCHEMA public;\n\n", quoteIdentifier(extname)); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpSchemas dumps CREATE SCHEMA statements for non-system schemas
func (s *PgDumpService) dumpSchemas() error {
	rows, err := s.db.Query(`
		SELECT nspname
		FROM pg_namespace
		WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		  AND nspname NOT LIKE 'pg_temp_%'
		  AND nspname NOT LIKE 'pg_toast_temp_%'
		ORDER BY nspname`)
	if err != nil {
		return fmt.Errorf("querying schemas: %w", err)
	}
	defer rows.Close()

	var hasSchemas bool
	for rows.Next() {
		var nspname string
		if err := rows.Scan(&nspname); err != nil {
			return fmt.Errorf("scanning schema: %w", err)
		}
		if !s.isSchemaIncluded(nspname) {
			continue
		}
		if nspname == "public" {
			continue // public schema always exists
		}
		if !hasSchemas {
			if err := s.writeln("\n--\n-- Schemas\n--\n"); err != nil {
				return err
			}
			hasSchemas = true
		}
		if s.options.Clean {
			if err := s.writef("DROP SCHEMA IF EXISTS %s CASCADE;\n", quoteIdentifier(nspname)); err != nil {
				return err
			}
		}
		if err := s.writef("CREATE SCHEMA %s;\n\n", quoteIdentifier(nspname)); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpEnumTypes dumps CREATE TYPE ... AS ENUM statements
func (s *PgDumpService) dumpEnumTypes() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, t.typname,
			array_agg(e.enumlabel ORDER BY e.enumsortorder) AS labels
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		JOIN pg_enum e ON t.oid = e.enumtypid
		WHERE t.typtype = 'e'
		GROUP BY n.nspname, t.typname
		ORDER BY n.nspname, t.typname`)
	if err != nil {
		return fmt.Errorf("querying enum types: %w", err)
	}
	defer rows.Close()

	var hasTypes bool
	for rows.Next() {
		var schema, typname string
		var labelsArr []byte // PostgreSQL array comes as text
		if err := rows.Scan(&schema, &typname, &labelsArr); err != nil {
			return fmt.Errorf("scanning enum type: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if !hasTypes {
			if err := s.writeln("\n--\n-- Enum Types\n--\n"); err != nil {
				return err
			}
			hasTypes = true
		}

		labels := parsePostgresArray(string(labelsArr))

		qname := qualifiedName(schema, typname)
		if s.options.Clean {
			if err := s.writef("DROP TYPE IF EXISTS %s CASCADE;\n", qname); err != nil {
				return err
			}
		}
		quotedLabels := make([]string, len(labels))
		for i, l := range labels {
			quotedLabels[i] = "'" + strings.ReplaceAll(l, "'", "''") + "'"
		}
		if err := s.writef("CREATE TYPE %s AS ENUM (\n    %s\n);\n\n", qname, strings.Join(quotedLabels, ",\n    ")); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpSequences dumps CREATE SEQUENCE statements
func (s *PgDumpService) dumpSequences() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'S'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY n.nspname, c.relname`)
	if err != nil {
		return fmt.Errorf("querying sequences: %w", err)
	}
	defer rows.Close()

	type seqInfo struct {
		schema, name string
	}
	var sequences []seqInfo
	for rows.Next() {
		var si seqInfo
		if err := rows.Scan(&si.schema, &si.name); err != nil {
			return fmt.Errorf("scanning sequence: %w", err)
		}
		if !s.isSchemaIncluded(si.schema) {
			continue
		}
		sequences = append(sequences, si)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(sequences) == 0 {
		return nil
	}

	if err := s.writeln("\n--\n-- Sequences\n--\n"); err != nil {
		return err
	}

	for _, seq := range sequences {
		qname := qualifiedName(seq.schema, seq.name)

		// Query sequence properties
		var startValue, minValue, maxValue, incrementBy, cacheValue int64
		var isCycled bool
		err := s.db.QueryRow(fmt.Sprintf(
			`SELECT start_value, min_value, max_value, increment_by, cache_size, cycle
			 FROM pg_sequences WHERE schemaname = $1 AND sequencename = $2`),
			seq.schema, seq.name).Scan(&startValue, &minValue, &maxValue, &incrementBy, &cacheValue, &isCycled)
		if err != nil {
			return fmt.Errorf("querying sequence properties for %s: %w", qname, err)
		}

		if s.options.Clean {
			if err := s.writef("DROP SEQUENCE IF EXISTS %s CASCADE;\n", qname); err != nil {
				return err
			}
		}

		cycleStr := "NO CYCLE"
		if isCycled {
			cycleStr = "CYCLE"
		}

		if err := s.writef("CREATE SEQUENCE %s\n    START WITH %d\n    INCREMENT BY %d\n    MINVALUE %d\n    MAXVALUE %d\n    CACHE %d\n    %s;\n\n",
			qname, startValue, incrementBy, minValue, maxValue, cacheValue, cycleStr); err != nil {
			return err
		}
	}
	return nil
}

// dumpTables dumps CREATE TABLE statements with columns, defaults, NOT NULL, primary keys, and unique constraints
func (s *PgDumpService) dumpTables() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'r'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND c.relname NOT LIKE 'pg_%'
		ORDER BY n.nspname, c.relname`)
	if err != nil {
		return fmt.Errorf("querying tables: %w", err)
	}
	defer rows.Close()

	type tableInfo struct {
		schema, name string
	}
	var tables []tableInfo
	for rows.Next() {
		var ti tableInfo
		if err := rows.Scan(&ti.schema, &ti.name); err != nil {
			return fmt.Errorf("scanning table: %w", err)
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

	if err := s.writeln("\n--\n-- Tables\n--\n"); err != nil {
		return err
	}

	for _, tbl := range tables {
		if err := s.dumpSingleTable(tbl.schema, tbl.name); err != nil {
			return fmt.Errorf("dumping table %s.%s: %w", tbl.schema, tbl.name, err)
		}
	}
	return nil
}

func (s *PgDumpService) dumpSingleTable(schema, table string) error {
	qname := qualifiedName(schema, table)

	if s.options.Clean {
		if err := s.writef("DROP TABLE IF EXISTS %s CASCADE;\n", qname); err != nil {
			return err
		}
	}

	// Get columns
	colRows, err := s.db.Query(`
		SELECT a.attname,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			a.attnotnull,
			pg_get_expr(d.adbin, d.adrelid) AS default_value
		FROM pg_attribute a
		LEFT JOIN pg_attrdef d ON a.attrelid = d.adrelid AND a.attnum = d.adnum
		WHERE a.attrelid = $1::regclass
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY a.attnum`, qname)
	if err != nil {
		return fmt.Errorf("querying columns: %w", err)
	}
	defer colRows.Close()

	var columns []string
	for colRows.Next() {
		var colName, dataType string
		var notNull bool
		var defaultValue sql.NullString
		if err := colRows.Scan(&colName, &dataType, &notNull, &defaultValue); err != nil {
			return fmt.Errorf("scanning column: %w", err)
		}
		col := fmt.Sprintf("    %s %s", quoteIdentifier(colName), dataType)
		if defaultValue.Valid {
			col += " DEFAULT " + defaultValue.String
		}
		if notNull {
			col += " NOT NULL"
		}
		columns = append(columns, col)
	}
	if err := colRows.Err(); err != nil {
		return err
	}

	// Get primary key and unique constraints
	conRows, err := s.db.Query(`
		SELECT conname, pg_get_constraintdef(oid)
		FROM pg_constraint
		WHERE conrelid = $1::regclass
		  AND contype IN ('p', 'u')
		ORDER BY contype, conname`, qname)
	if err != nil {
		return fmt.Errorf("querying constraints: %w", err)
	}
	defer conRows.Close()

	var constraints []string
	for conRows.Next() {
		var conname, condef string
		if err := conRows.Scan(&conname, &condef); err != nil {
			return fmt.Errorf("scanning constraint: %w", err)
		}
		constraints = append(constraints, fmt.Sprintf("    CONSTRAINT %s %s", quoteIdentifier(conname), condef))
	}
	if err := conRows.Err(); err != nil {
		return err
	}

	// Build CREATE TABLE
	allParts := append(columns, constraints...)
	if err := s.writef("CREATE TABLE %s (\n%s\n);\n\n", qname, strings.Join(allParts, ",\n")); err != nil {
		return err
	}

	return nil
}

// dumpViews dumps CREATE VIEW statements
func (s *PgDumpService) dumpViews() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname, pg_get_viewdef(c.oid, true) AS viewdef
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'v'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY n.nspname, c.relname`)
	if err != nil {
		return fmt.Errorf("querying views: %w", err)
	}
	defer rows.Close()

	var hasViews bool
	for rows.Next() {
		var schema, name, viewdef string
		if err := rows.Scan(&schema, &name, &viewdef); err != nil {
			return fmt.Errorf("scanning view: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if !hasViews {
			if err := s.writeln("\n--\n-- Views\n--\n"); err != nil {
				return err
			}
			hasViews = true
		}
		qname := qualifiedName(schema, name)
		if s.options.Clean {
			if err := s.writef("DROP VIEW IF EXISTS %s CASCADE;\n", qname); err != nil {
				return err
			}
		}
		if err := s.writef("CREATE VIEW %s AS\n%s;\n\n", qname, strings.TrimRight(viewdef, " \t\n;")); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpFunctions dumps CREATE FUNCTION statements
func (s *PgDumpService) dumpFunctions() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, p.proname, pg_get_functiondef(p.oid) AS funcdef
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND p.prokind != 'a'
		ORDER BY n.nspname, p.proname`)
	if err != nil {
		return fmt.Errorf("querying functions: %w", err)
	}
	defer rows.Close()

	var hasFunctions bool
	for rows.Next() {
		var schema, name, funcdef string
		if err := rows.Scan(&schema, &name, &funcdef); err != nil {
			return fmt.Errorf("scanning function: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if !hasFunctions {
			if err := s.writeln("\n--\n-- Functions\n--\n"); err != nil {
				return err
			}
			hasFunctions = true
		}
		if s.options.Clean {
			// Use the function definition to get the full signature for DROP
			if err := s.writef("-- (clean mode: function will be replaced)\n"); err != nil {
				return err
			}
		}
		if err := s.writef("%s;\n\n", strings.TrimRight(funcdef, " \t\n;")); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpIndexes dumps CREATE INDEX statements (non-PK, non-unique-constraint indexes)
func (s *PgDumpService) dumpIndexes() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname AS tablename, i.relname AS indexname,
			pg_get_indexdef(i.oid) AS indexdef
		FROM pg_index x
		JOIN pg_class c ON c.oid = x.indrelid
		JOIN pg_class i ON i.oid = x.indexrelid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND NOT x.indisprimary
		  AND NOT x.indisunique
		ORDER BY n.nspname, c.relname, i.relname`)
	if err != nil {
		return fmt.Errorf("querying indexes: %w", err)
	}
	defer rows.Close()

	var hasIndexes bool
	for rows.Next() {
		var schema, tablename, indexname, indexdef string
		if err := rows.Scan(&schema, &tablename, &indexname, &indexdef); err != nil {
			return fmt.Errorf("scanning index: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if s.isTableExcluded(tablename) {
			continue
		}
		if !hasIndexes {
			if err := s.writeln("\n--\n-- Indexes\n--\n"); err != nil {
				return err
			}
			hasIndexes = true
		}
		if s.options.Clean {
			if err := s.writef("DROP INDEX IF EXISTS %s CASCADE;\n", qualifiedName(schema, indexname)); err != nil {
				return err
			}
		}
		if err := s.writef("%s;\n\n", indexdef); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpForeignKeys dumps ALTER TABLE ... ADD CONSTRAINT ... FOREIGN KEY statements
func (s *PgDumpService) dumpForeignKeys() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname, con.conname, pg_get_constraintdef(con.oid)
		FROM pg_constraint con
		JOIN pg_class c ON con.conrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE con.contype = 'f'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY n.nspname, c.relname, con.conname`)
	if err != nil {
		return fmt.Errorf("querying foreign keys: %w", err)
	}
	defer rows.Close()

	var hasFKs bool
	for rows.Next() {
		var schema, tablename, conname, condef string
		if err := rows.Scan(&schema, &tablename, &conname, &condef); err != nil {
			return fmt.Errorf("scanning foreign key: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if s.isTableExcluded(tablename) {
			continue
		}
		if !hasFKs {
			if err := s.writeln("\n--\n-- Foreign Keys\n--\n"); err != nil {
				return err
			}
			hasFKs = true
		}
		qname := qualifiedName(schema, tablename)
		if err := s.writef("ALTER TABLE %s ADD CONSTRAINT %s %s;\n\n",
			qname, quoteIdentifier(conname), condef); err != nil {
			return err
		}
	}
	return rows.Err()
}

// dumpTriggers dumps CREATE TRIGGER statements
func (s *PgDumpService) dumpTriggers() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname, t.tgname, pg_get_triggerdef(t.oid) AS triggerdef
		FROM pg_trigger t
		JOIN pg_class c ON t.tgrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE NOT t.tgisinternal
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY n.nspname, c.relname, t.tgname`)
	if err != nil {
		return fmt.Errorf("querying triggers: %w", err)
	}
	defer rows.Close()

	var hasTriggers bool
	for rows.Next() {
		var schema, tablename, tgname, triggerdef string
		if err := rows.Scan(&schema, &tablename, &tgname, &triggerdef); err != nil {
			return fmt.Errorf("scanning trigger: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if s.isTableExcluded(tablename) {
			continue
		}
		if !hasTriggers {
			if err := s.writeln("\n--\n-- Triggers\n--\n"); err != nil {
				return err
			}
			hasTriggers = true
		}
		if s.options.Clean {
			if err := s.writef("DROP TRIGGER IF EXISTS %s ON %s CASCADE;\n",
				quoteIdentifier(tgname), qualifiedName(schema, tablename)); err != nil {
				return err
			}
		}
		if err := s.writef("%s;\n\n", triggerdef); err != nil {
			return err
		}
	}
	return rows.Err()
}

// parsePostgresArray parses a PostgreSQL text array literal like {a,b,c}
func parsePostgresArray(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		return nil
	}
	// Remove surrounding braces
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	var result []string
	var current strings.Builder
	inQuote := false
	escaped := false

	for _, r := range s {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '"' {
			inQuote = !inQuote
			continue
		}
		if r == ',' && !inQuote {
			result = append(result, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
