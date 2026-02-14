package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// PgDumpBinaryDumper uses the external pg_dump binary to dump a database
type PgDumpBinaryDumper struct {
	connParams *models.ConnectionParameters
	options    PgDumpOptions
}

// NewPgDumpBinaryDumper creates a new pg_dump binary dumper
func NewPgDumpBinaryDumper(connParams *models.ConnectionParameters, options PgDumpOptions) *PgDumpBinaryDumper {
	return &PgDumpBinaryDumper{
		connParams: connParams,
		options:    options,
	}
}

// Dump executes pg_dump and writes output to the provided writer
func (d *PgDumpBinaryDumper) Dump(w io.Writer) error {
	var args []string

	// pg_dump accepts a conninfo string as the --dbname argument
	if d.connParams.RawConnectionString != "" {
		args = []string{
			fmt.Sprintf("--dbname=%s", d.connParams.RawConnectionString),
			"--format=plain",
		}
	} else {
		args = []string{
			fmt.Sprintf("--host=%s", d.connParams.Host),
			fmt.Sprintf("--port=%d", d.connParams.Port),
			fmt.Sprintf("--username=%s", d.connParams.User),
			fmt.Sprintf("--dbname=%s", d.connParams.DatabaseName),
			"--format=plain",
		}
	}

	if d.options.NoOwner {
		args = append(args, "--no-owner", "--no-acl")
	}
	if d.options.Clean {
		args = append(args, "--clean", "--if-exists")
	}
	for _, schema := range d.options.Schemas {
		args = append(args, fmt.Sprintf("--schema=%s", schema))
	}
	for _, table := range d.options.ExcludeTables {
		args = append(args, fmt.Sprintf("--exclude-table=%s", table))
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if d.connParams.RawConnectionString == "" && d.connParams.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.connParams.Password))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	return nil
}
