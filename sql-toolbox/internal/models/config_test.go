package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_MigrationOnly(t *testing.T) {
	yamlContent := `
migration:
  target_path: "./sql-scripts"
  provider: "postgresql"
  host: "localhost"
  port: 5432
  database: "myapp"
  user: "admin"
  password: "secret"
  backup: true
`
	path := writeTempConfig(t, yamlContent)
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Migration == nil {
		t.Fatal("expected migration config to be non-nil")
	}
	if config.Migration.TargetPath != "./sql-scripts" {
		t.Errorf("expected target_path './sql-scripts', got '%s'", config.Migration.TargetPath)
	}
	if config.Migration.Provider != "postgresql" {
		t.Errorf("expected provider 'postgresql', got '%s'", config.Migration.Provider)
	}
	if config.Migration.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", config.Migration.Host)
	}
	if config.Migration.Port != 5432 {
		t.Errorf("expected port 5432, got %d", config.Migration.Port)
	}
	if !config.Migration.Backup {
		t.Error("expected backup to be true")
	}
	if config.Backup != nil {
		t.Error("expected backup config to be nil")
	}
}

func TestLoadConfig_BackupOnly(t *testing.T) {
	yamlContent := `
backup:
  defaults:
    output_dir: "./backups"
    compress: true
  databases:
    - name: "myapp_prod"
      host: "db.example.com"
      port: 5432
      database: "myapp"
      user: "backup_user"
      password: "secret"
      ssl_mode: "disable"
  s3:
    enabled: false
    bucket: "my-backups"
    region: "us-east-1"
`
	path := writeTempConfig(t, yamlContent)
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Migration != nil {
		t.Error("expected migration config to be nil")
	}
	if config.Backup == nil {
		t.Fatal("expected backup config to be non-nil")
	}
	if len(config.Backup.Databases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(config.Backup.Databases))
	}
	if config.Backup.Databases[0].Name != "myapp_prod" {
		t.Errorf("expected name 'myapp_prod', got '%s'", config.Backup.Databases[0].Name)
	}
}

func TestLoadConfig_UnifiedConfig(t *testing.T) {
	yamlContent := `
migration:
  target_path: "./sql-scripts"
  provider: "postgresql"
  host: "localhost"
  port: 5432
  database: "myapp"
  user: "admin"
  password: "secret"

backup:
  defaults:
    output_dir: "./backups"
  databases:
    - name: "myapp_prod"
      host: "db.example.com"
      port: 5432
      database: "myapp"
      user: "backup_user"
      password: "secret"
`
	path := writeTempConfig(t, yamlContent)
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Migration == nil {
		t.Fatal("expected migration config to be non-nil")
	}
	if config.Backup == nil {
		t.Fatal("expected backup config to be non-nil")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	path := writeTempConfig(t, "")
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Migration != nil {
		t.Error("expected migration config to be nil for empty file")
	}
	if config.Backup != nil {
		t.Error("expected backup config to be nil for empty file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	path := writeTempConfig(t, "{{invalid yaml")
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_MigrationBackupMethod(t *testing.T) {
	yamlContent := `
migration:
  target_path: "./scripts"
  provider: "postgresql"
  host: "localhost"
  port: 5432
  database: "testdb"
  backup_method: "go"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Migration == nil {
		t.Fatal("expected migration config")
	}
	if config.Migration.BackupMethod != "go" {
		t.Errorf("expected backup_method 'go', got '%s'", config.Migration.BackupMethod)
	}
}

func TestLoadConfig_MigrationDefaults(t *testing.T) {
	yamlContent := `
migration:
  database: "myapp"
`
	path := writeTempConfig(t, yamlContent)
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Migration.Provider != "" {
		t.Errorf("expected empty provider, got '%s'", config.Migration.Provider)
	}
	if config.Migration.Port != 0 {
		t.Errorf("expected port 0, got %d", config.Migration.Port)
	}
	if config.Migration.Backup {
		t.Error("expected backup to be false by default")
	}
}

