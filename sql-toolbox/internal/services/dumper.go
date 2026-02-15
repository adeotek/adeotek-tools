package services

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// BackupMethodPgDump uses the external pg_dump binary
const BackupMethodPgDump = "pg_dump"

// BackupMethodGo uses the pure Go dump implementation
const BackupMethodGo = "go"

// DatabaseDumper defines the interface for database dump implementations
type DatabaseDumper interface {
	Dump(w io.Writer) error
}

// NewDatabaseDumper creates the appropriate dumper based on the method
func NewDatabaseDumper(method string, connParams *models.ConnectionParameters, db *sql.DB, options PgDumpOptions) (DatabaseDumper, error) {
	switch method {
	case BackupMethodPgDump:
		if connParams == nil {
			return nil, fmt.Errorf("connection parameters required for pg_dump method")
		}
		return NewPgDumpBinaryDumper(connParams, options), nil
	case BackupMethodGo:
		if db == nil {
			return nil, fmt.Errorf("database connection required for go method")
		}
		return NewPgDumpGoDumper(db, options), nil
	default:
		return nil, fmt.Errorf("unsupported backup method: %s", method)
	}
}
