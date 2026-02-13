package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

func TestBackupOnlyOrchestrator_DryRun(t *testing.T) {
	config := &models.BackupConfig{
		Defaults: models.BackupDefaults{
			OutputDir: t.TempDir(),
		},
		Databases: []models.DatabaseTarget{
			{
				Name:             "testdb",
				Host:             "localhost",
				Port:             5432,
				Database:         "testdb",
				User:             "user",
				Password:         "pass",
				ConnectionString: "",
			},
		},
	}
	// Need to call LoadBackupConfig-like linking
	linkConfigDefaults(config)

	orchestrator := NewBackupOnlyOrchestrator(config, true, true)
	summary, err := orchestrator.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 1 {
		t.Errorf("expected total 1, got %d", summary.Total)
	}
	if summary.Succeeded != 1 {
		t.Errorf("expected succeeded 1, got %d", summary.Succeeded)
	}
	if summary.Failed != 0 {
		t.Errorf("expected failed 0, got %d", summary.Failed)
	}
}

func TestBackupOnlyOrchestrator_DryRunWithS3(t *testing.T) {
	config := &models.BackupConfig{
		Defaults: models.BackupDefaults{
			OutputDir:  t.TempDir(),
			UploadToS3: boolP(true),
		},
		Databases: []models.DatabaseTarget{
			{
				Name:     "testdb",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
		},
		S3: models.S3Config{
			Enabled: true,
			Bucket:  "test-bucket",
			Region:  "us-east-1",
		},
	}
	linkConfigDefaults(config)

	mock := &MockS3Uploader{}
	orchestrator := NewBackupOnlyOrchestrator(config, false, true)
	orchestrator.s3UploaderFactory = func(_ models.S3Config) (S3Uploader, error) {
		return mock, nil
	}

	summary, err := orchestrator.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Succeeded != 1 {
		t.Errorf("expected succeeded 1, got %d", summary.Succeeded)
	}
	// In dry-run mode, no actual upload should happen
	if len(mock.UploadCalls) != 0 {
		t.Errorf("expected 0 upload calls in dry-run, got %d", len(mock.UploadCalls))
	}
	if !summary.Results[0].UploadedToS3 {
		t.Error("expected UploadedToS3 to be true in dry-run")
	}
}

func TestBackupOnlyOrchestrator_ConnectionFailure(t *testing.T) {
	config := &models.BackupConfig{
		Defaults: models.BackupDefaults{
			OutputDir: t.TempDir(),
		},
		Databases: []models.DatabaseTarget{
			{
				Name:     "faildb",
				Host:     "nonexistent",
				Port:     5432,
				Database: "faildb",
			},
		},
	}
	linkConfigDefaults(config)

	orchestrator := NewBackupOnlyOrchestrator(config, false, false)
	// Override connector to simulate failure
	orchestrator.dbConnector = func(connStr string) (*sql.DB, error) {
		return nil, fmt.Errorf("connection refused")
	}

	summary, err := orchestrator.Run()
	if err != nil {
		t.Fatalf("unexpected error from Run: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", summary.Failed)
	}
	if summary.Results[0].Error == nil {
		t.Error("expected error in result")
	}
}

func TestBackupOnlyOrchestrator_PartialFailure(t *testing.T) {
	config := &models.BackupConfig{
		Defaults: models.BackupDefaults{
			OutputDir: t.TempDir(),
		},
		Databases: []models.DatabaseTarget{
			{Name: "db1", Host: "localhost", Port: 5432, Database: "db1"},
			{Name: "db2", Host: "nonexistent", Port: 5432, Database: "db2"},
		},
	}
	linkConfigDefaults(config)

	callCount := 0
	orchestrator := NewBackupOnlyOrchestrator(config, false, false)
	orchestrator.dbConnector = func(connStr string) (*sql.DB, error) {
		callCount++
		if callCount == 2 {
			return nil, fmt.Errorf("connection refused")
		}
		// First call also fails since we can't create a real DB connection in tests
		return nil, fmt.Errorf("connection refused")
	}

	summary, err := orchestrator.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 2 {
		t.Errorf("expected total 2, got %d", summary.Total)
	}
	if summary.Failed != 2 {
		t.Errorf("expected 2 failures, got %d", summary.Failed)
	}
}

func TestCreateOutputWriter_Plain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sql")

	writer, closer, err := createOutputWriter(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = writer.Write([]byte("SELECT 1;"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := closer(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if string(data) != "SELECT 1;" {
		t.Errorf("expected 'SELECT 1;', got '%s'", string(data))
	}
}

func TestCreateOutputWriter_Gzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sql.gz")

	writer, closer, err := createOutputWriter(path, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = writer.Write([]byte("SELECT 1;"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := closer(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	// Verify file exists and is non-empty (gzip header makes it larger than 0)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty gzip file")
	}
}

// linkConfigDefaults links defaults and S3 config to database targets (mimics LoadBackupConfig)
func linkConfigDefaults(config *models.BackupConfig) {
	for i := range config.Databases {
		config.Databases[i].SetDefaults(&config.Defaults, &config.S3)
	}
}

func boolP(b bool) *bool {
	return &b
}
