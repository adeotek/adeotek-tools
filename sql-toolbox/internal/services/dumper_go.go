package services

import (
	"database/sql"
	"fmt"
	"io"
)

// PgDumpGoDumper wraps the existing PgDumpService as a DatabaseDumper
type PgDumpGoDumper struct {
	db      *sql.DB
	options PgDumpOptions
}

// NewPgDumpGoDumper creates a new pure Go dumper
func NewPgDumpGoDumper(db *sql.DB, options PgDumpOptions) *PgDumpGoDumper {
	return &PgDumpGoDumper{
		db:      db,
		options: options,
	}
}

// Dump performs the database dump using the pure Go implementation
func (d *PgDumpGoDumper) Dump(w io.Writer) error {
	if d.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	service := NewPgDumpService(d.db, d.options, w)
	return service.Dump()
}
