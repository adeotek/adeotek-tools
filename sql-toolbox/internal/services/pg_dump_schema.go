package services

import (
	"database/sql"
	"fmt"
	"strings"
)

// dumpExtensions dumps CREATE EXTENSION statements
func (s *PgDumpService) dumpExtensions() error {
	rows, err := s.db.Query(`
		SELECT e.extname, n.nspname
		FROM pg_extension e
		JOIN pg_namespace n ON n.oid = e.extnamespace
		WHERE e.extname != 'plpgsql'
		ORDER BY e.extname`)
	if err != nil {
		return fmt.Errorf("querying extensions: %w", err)
	}
	defer rows.Close()

	var hasExtensions bool
	for rows.Next() {
		var extname, schema string
		if err := rows.Scan(&extname, &schema); err != nil {
			return fmt.Errorf("scanning extension: %w", err)
		}
		if !hasExtensions {
			if err := s.writeln("\n--\n-- Extensions\n--\n"); err != nil {
				return err
			}
			hasExtensions = true
		}
		if s.options.Clean {
			if err := s.writef("DROP EXTENSION IF EXISTS %s CASCADE;\n", quoteIdentifier(extname)); err != nil {
				return err
			}
		}
		if err := s.writef("CREATE EXTENSION IF NOT EXISTS %s WITH SCHEMA %s;\n\n",
			quoteIdentifier(extname), quoteIdentifier(schema)); err != nil {
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
		  AND NOT EXISTS (
			SELECT 1 FROM pg_depend d
			WHERE d.classid = 'pg_type'::regclass AND d.objid = t.oid AND d.deptype = 'e')
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
		  AND NOT EXISTS (
			SELECT 1 FROM pg_depend d
			WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e')
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
		err := s.db.QueryRow(
			`SELECT start_value, min_value, max_value, increment_by, cache_size, cycle
			 FROM pg_sequences WHERE schemaname = $1 AND sequencename = $2`,
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

		// Preserve the current sequence position, otherwise restored sequences
		// would hand out values that are already in use.
		var lastValue int64
		var isCalled bool
		if err := s.db.QueryRow(fmt.Sprintf("SELECT last_value, is_called FROM %s", qname)).
			Scan(&lastValue, &isCalled); err != nil {
			return fmt.Errorf("querying sequence state for %s: %w", qname, err)
		}
		if err := s.writef("SELECT pg_catalog.setval(%s, %d, %t);\n\n",
			quoteLiteral(qname), lastValue, isCalled); err != nil {
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
		  AND NOT EXISTS (
			SELECT 1 FROM pg_depend d
			WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e')
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
		if s.isTableExcluded(ti.schema, ti.name) {
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
			pg_get_expr(d.adbin, d.adrelid) AS default_value,
			a.attidentity,
			a.attgenerated
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
		var colName, dataType, identity, generated string
		var notNull bool
		var defaultValue sql.NullString
		if err := colRows.Scan(&colName, &dataType, &notNull, &defaultValue, &identity, &generated); err != nil {
			return fmt.Errorf("scanning column: %w", err)
		}
		col := fmt.Sprintf("    %s %s", quoteIdentifier(colName), dataType)
		switch {
		case generated != "":
			// Generated columns carry their expression in pg_attrdef; a DEFAULT
			// clause would make the CREATE TABLE invalid.
			if defaultValue.Valid {
				col += fmt.Sprintf(" GENERATED ALWAYS AS (%s) STORED", defaultValue.String)
			}
		case identity == "a":
			col += " GENERATED ALWAYS AS IDENTITY"
		case identity == "d":
			col += " GENERATED BY DEFAULT AS IDENTITY"
		case defaultValue.Valid:
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
		  AND contype IN ('p', 'u', 'c', 'x')
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

type viewDefinition struct {
	schema       string
	name         string
	relkind      string
	definition   string
	dependencies []string
}

func (v viewDefinition) key() string {
	return v.schema + "." + v.name
}

// queryViews returns views and materialized views with their view dependencies.
func (s *PgDumpService) queryViews() ([]viewDefinition, error) {
	rows, err := s.db.Query(`
		SELECT n.nspname, c.relname, c.relkind, pg_get_viewdef(c.oid, true) AS viewdef,
			COALESCE(ARRAY(
				SELECT DISTINCT dn.nspname || '.' || dc.relname
				FROM pg_rewrite rw
				JOIN pg_depend d ON d.objid = rw.oid
				JOIN pg_class dc ON dc.oid = d.refobjid
				JOIN pg_namespace dn ON dn.oid = dc.relnamespace
				WHERE rw.ev_class = c.oid
				  AND d.refobjid <> c.oid
				  AND dc.relkind IN ('v', 'm')
			), '{}') AS deps
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind IN ('v', 'm')
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY n.nspname, c.relname`)
	if err != nil {
		return nil, fmt.Errorf("querying views: %w", err)
	}
	defer rows.Close()

	var views []viewDefinition
	for rows.Next() {
		var view viewDefinition
		var deps []byte
		if err := rows.Scan(&view.schema, &view.name, &view.relkind, &view.definition, &deps); err != nil {
			return nil, fmt.Errorf("scanning view: %w", err)
		}
		if !s.isSchemaIncluded(view.schema) {
			continue
		}
		view.dependencies = parsePostgresArray(string(deps))
		views = append(views, view)
	}
	return views, rows.Err()
}

// orderViews sorts views so that dependencies are created before dependents.
func orderViews(views []viewDefinition) (ordered, unresolved []viewDefinition) {
	known := make(map[string]bool, len(views))
	for _, view := range views {
		known[view.key()] = true
	}

	pending := views
	emitted := make(map[string]bool, len(views))
	for len(pending) > 0 {
		progressed := false
		var remaining []viewDefinition
		for _, view := range pending {
			ready := true
			for _, dep := range view.dependencies {
				if known[dep] && !emitted[dep] {
					ready = false
					break
				}
			}
			if ready {
				ordered = append(ordered, view)
				emitted[view.key()] = true
				progressed = true
			} else {
				remaining = append(remaining, view)
			}
		}
		if !progressed {
			// Circular or unresolvable dependency information: emit the rest in
			// their original order rather than dropping them.
			return ordered, remaining
		}
		pending = remaining
	}
	return ordered, nil
}

// dumpViews dumps CREATE VIEW / CREATE MATERIALIZED VIEW statements
func (s *PgDumpService) dumpViews() error {
	views, err := s.queryViews()
	if err != nil {
		return err
	}
	if len(views) == 0 {
		return nil
	}

	ordered, unresolved := orderViews(views)
	if len(unresolved) > 0 {
		if err := s.writeln("-- WARNING: could not resolve view dependency order for:"); err != nil {
			return err
		}
		for _, view := range unresolved {
			if err := s.writef("--   %s\n", view.key()); err != nil {
				return err
			}
		}
		ordered = append(ordered, unresolved...)
	}

	if err := s.writeln("\n--\n-- Views\n--\n"); err != nil {
		return err
	}
	for _, view := range ordered {
		qname := qualifiedName(view.schema, view.name)
		kind := "VIEW"
		if view.relkind == "m" {
			kind = "MATERIALIZED VIEW"
		}
		if s.options.Clean {
			if err := s.writef("DROP %s IF EXISTS %s CASCADE;\n", kind, qname); err != nil {
				return err
			}
		}
		suffix := ""
		if view.relkind == "m" {
			suffix = " WITH NO DATA"
		}
		if err := s.writef("CREATE %s %s AS\n%s%s;\n\n",
			kind, qname, strings.TrimRight(view.definition, " \t\n;"), suffix); err != nil {
			return err
		}
	}
	return nil
}

// dumpMaterializedViewData refreshes materialized views after base data is loaded.
func (s *PgDumpService) dumpMaterializedViewData() error {
	views, err := s.queryViews()
	if err != nil {
		return err
	}
	var wroteHeader bool
	for _, view := range views {
		if view.relkind != "m" {
			continue
		}
		if !wroteHeader {
			if err := s.writeln("\n--\n-- Materialized view data\n--\n"); err != nil {
				return err
			}
			wroteHeader = true
		}
		if err := s.writef("REFRESH MATERIALIZED VIEW %s;\n", qualifiedName(view.schema, view.name)); err != nil {
			return err
		}
	}
	return nil
}

// dumpFunctions dumps CREATE FUNCTION statements
func (s *PgDumpService) dumpFunctions() error {
	rows, err := s.db.Query(`
		SELECT n.nspname, p.proname, pg_get_functiondef(p.oid) AS funcdef
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND p.prokind != 'a'
		  AND NOT EXISTS (
			SELECT 1 FROM pg_depend d
			WHERE d.classid = 'pg_proc'::regclass AND d.objid = p.oid AND d.deptype = 'e')
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
		  AND NOT EXISTS (
			SELECT 1 FROM pg_constraint con
			WHERE con.conindid = x.indexrelid)
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
		if s.isTableExcluded(schema, tablename) {
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
		SELECT n.nspname, c.relname, con.conname, pg_get_constraintdef(con.oid),
			rn.nspname, rc.relname
		FROM pg_constraint con
		JOIN pg_class c ON con.conrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		JOIN pg_class rc ON con.confrelid = rc.oid
		JOIN pg_namespace rn ON rc.relnamespace = rn.oid
		WHERE con.contype = 'f'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY n.nspname, c.relname, con.conname`)
	if err != nil {
		return fmt.Errorf("querying foreign keys: %w", err)
	}
	defer rows.Close()

	var hasFKs bool
	for rows.Next() {
		var schema, tablename, conname, condef, refSchema, refTable string
		if err := rows.Scan(&schema, &tablename, &conname, &condef, &refSchema, &refTable); err != nil {
			return fmt.Errorf("scanning foreign key: %w", err)
		}
		if !s.isSchemaIncluded(schema) {
			continue
		}
		if s.isTableExcluded(schema, tablename) || s.isTableExcluded(refSchema, refTable) {
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
		if s.isTableExcluded(schema, tablename) {
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
