package services

import (
	"bytes"
	"testing"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

func TestNewDatabaseDumper_InvalidMethod(t *testing.T) {
	_, err := NewDatabaseDumper("invalid", nil, nil, PgDumpOptions{})
	if err == nil {
		t.Error("expected error for invalid method")
	}
}

func TestNewDatabaseDumper_GoMethodRequiresDB(t *testing.T) {
	_, err := NewDatabaseDumper("go", nil, nil, PgDumpOptions{})
	if err == nil {
		t.Error("expected error when db is nil for go method")
	}
}

func TestNewDatabaseDumper_PgDumpMethodRequiresConnParams(t *testing.T) {
	_, err := NewDatabaseDumper("pg_dump", nil, nil, PgDumpOptions{})
	if err == nil {
		t.Error("expected error when connParams is nil for pg_dump method")
	}
}

func TestNewDatabaseDumper_PgDumpMethodCreatesInstance(t *testing.T) {
	connParams := &models.ConnectionParameters{
		Provider:     "postgresql",
		Host:         "localhost",
		Port:         5432,
		DatabaseName: "testdb",
		User:         "user",
		Password:     "pass",
	}
	dumper, err := NewDatabaseDumper("pg_dump", connParams, nil, PgDumpOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dumper.(*PgDumpBinaryDumper); !ok {
		t.Error("expected PgDumpBinaryDumper instance")
	}
}

func TestPgDumpGoDumper_NilDB(t *testing.T) {
	dumper := &PgDumpGoDumper{db: nil, options: PgDumpOptions{}}
	var buf bytes.Buffer
	err := dumper.Dump(&buf)
	if err == nil {
		t.Error("expected error when db is nil")
	}
}
