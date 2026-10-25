package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// A dry run must never create a SQLite database file (checking the migration history
// used to open a connection, which created the file as a side effect).
func TestMigrationService_DryRun_DoesNotCreateSQLiteDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "dry-run.db")

	scriptsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(scriptsDir, "001_create.sql"), []byte("CREATE TABLE t (id INTEGER);\n"), 0644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	params := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: dbPath,
	}

	service := NewMigrationService(true, false, nil)
	if err := service.Run(scriptsDir, params); err != nil {
		t.Fatalf("dry-run migration failed: %v", err)
	}

	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Errorf("dry run created the SQLite database file, stat err: %v", err)
	}
}
