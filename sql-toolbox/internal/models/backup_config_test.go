package models

import (
	"os"
	"path/filepath"
	"testing"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestLoadBackupConfig_ValidConfig(t *testing.T) {
	yaml := `
defaults:
  output_dir: "./backups"
  compress: true
  upload_to_s3: false
  no_owner: true
  clean: true
  schemas: ["public"]

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
	path := writeTempConfig(t, yaml)
	config, err := LoadBackupConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Databases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(config.Databases))
	}
	if config.Databases[0].Name != "myapp_prod" {
		t.Errorf("expected name 'myapp_prod', got '%s'", config.Databases[0].Name)
	}
	if config.Defaults.OutputDir != "./backups" {
		t.Errorf("expected output_dir './backups', got '%s'", config.Defaults.OutputDir)
	}
}

func TestLoadBackupConfig_NoDatabases(t *testing.T) {
	yaml := `
defaults:
  output_dir: "./backups"
databases: []
`
	path := writeTempConfig(t, yaml)
	_, err := LoadBackupConfig(path)
	if err == nil {
		t.Fatal("expected error for empty databases")
	}
}

func TestLoadBackupConfig_MissingName(t *testing.T) {
	yaml := `
databases:
  - host: "db.example.com"
    database: "myapp"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadBackupConfig(path)
	if err == nil {
		t.Fatal("expected error for missing database name")
	}
}

func TestLoadBackupConfig_MissingHostAndConnectionString(t *testing.T) {
	yaml := `
databases:
  - name: "myapp"
    database: "myapp"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadBackupConfig(path)
	if err == nil {
		t.Fatal("expected error for missing host and connection_string")
	}
}

func TestLoadBackupConfig_ConnectionString(t *testing.T) {
	yaml := `
databases:
  - name: "myapp"
    connection_string: "host=localhost port=5432 dbname=myapp user=admin password=pass sslmode=disable"
`
	path := writeTempConfig(t, yaml)
	config, err := LoadBackupConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "host=localhost port=5432 dbname=myapp user=admin password=pass sslmode=disable"
	if config.Databases[0].ToConnectionString() != expected {
		t.Errorf("expected connection string '%s', got '%s'", expected, config.Databases[0].ToConnectionString())
	}
}

func TestLoadBackupConfig_S3EnabledNoBucket(t *testing.T) {
	yaml := `
databases:
  - name: "myapp"
    host: "localhost"
    database: "myapp"
s3:
  enabled: true
`
	path := writeTempConfig(t, yaml)
	_, err := LoadBackupConfig(path)
	if err == nil {
		t.Fatal("expected error for S3 enabled without bucket")
	}
}

func TestDatabaseTarget_GetEffectiveOutputDir(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name:     "per-db override",
			db:       DatabaseTarget{OutputDir: "/custom"},
			expected: "/custom",
		},
		{
			name:     "defaults fallback",
			db:       DatabaseTarget{defaults: &BackupDefaults{OutputDir: "/defaults"}},
			expected: "/defaults",
		},
		{
			name:     "hardcoded default",
			db:       DatabaseTarget{defaults: &BackupDefaults{}},
			expected: "./backups",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveOutputDir()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveCompress(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected bool
	}{
		{
			name:     "per-db true",
			db:       DatabaseTarget{Compress: boolPtr(true)},
			expected: true,
		},
		{
			name:     "per-db false",
			db:       DatabaseTarget{Compress: boolPtr(false)},
			expected: false,
		},
		{
			name:     "defaults fallback false",
			db:       DatabaseTarget{defaults: &BackupDefaults{Compress: boolPtr(false)}},
			expected: false,
		},
		{
			name:     "hardcoded default true",
			db:       DatabaseTarget{defaults: &BackupDefaults{}},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveCompress()
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveUploadToS3(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected bool
	}{
		{
			name:     "per-db true",
			db:       DatabaseTarget{UploadToS3: boolPtr(true)},
			expected: true,
		},
		{
			name:     "defaults fallback true",
			db:       DatabaseTarget{defaults: &BackupDefaults{UploadToS3: boolPtr(true)}},
			expected: true,
		},
		{
			name:     "hardcoded default false",
			db:       DatabaseTarget{defaults: &BackupDefaults{}},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveUploadToS3()
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveDeleteLocalAfterUpload(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected bool
	}{
		{
			name:     "per-db true",
			db:       DatabaseTarget{DeleteLocalAfterUpload: boolPtr(true)},
			expected: true,
		},
		{
			name:     "falls back to s3 config",
			db:       DatabaseTarget{s3Config: &S3Config{DeleteLocalAfterUpload: boolPtr(true)}},
			expected: true,
		},
		{
			name:     "hardcoded default false",
			db:       DatabaseTarget{s3Config: &S3Config{}},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveDeleteLocalAfterUpload()
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveSchemas(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected []string
	}{
		{
			name:     "per-db schemas",
			db:       DatabaseTarget{Schemas: []string{"public", "audit"}},
			expected: []string{"public", "audit"},
		},
		{
			name:     "defaults fallback",
			db:       DatabaseTarget{defaults: &BackupDefaults{Schemas: []string{"public"}}},
			expected: []string{"public"},
		},
		{
			name:     "nil when not set",
			db:       DatabaseTarget{defaults: &BackupDefaults{}},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveSchemas()
			if len(got) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, got)
					break
				}
			}
		})
	}
}

func TestDatabaseTarget_ToConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name: "from individual params",
			db: DatabaseTarget{
				Host:     "db.example.com",
				Port:     5432,
				Database: "myapp",
				User:     "admin",
				Password: "secret",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5432 dbname=myapp user=admin password=secret sslmode=require",
		},
		{
			name: "default port and sslmode",
			db: DatabaseTarget{
				Host:     "localhost",
				Database: "testdb",
				User:     "user",
			},
			expected: "host=localhost port=5432 dbname=testdb user=user sslmode=disable",
		},
		{
			name: "connection string takes precedence",
			db: DatabaseTarget{
				ConnectionString: "host=custom dbname=custom",
				Host:             "other",
				Database:         "other",
			},
			expected: "host=custom dbname=custom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.ToConnectionString()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestDatabaseTarget_GetEffectiveBackupMethod(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name:     "per-db override pg_dump",
			db:       DatabaseTarget{BackupMethod: strPtr("pg_dump")},
			expected: "pg_dump",
		},
		{
			name:     "per-db override go",
			db:       DatabaseTarget{BackupMethod: strPtr("go")},
			expected: "go",
		},
		{
			name:     "defaults fallback pg_dump",
			db:       DatabaseTarget{defaults: &BackupDefaults{BackupMethod: strPtr("pg_dump")}},
			expected: "pg_dump",
		},
		{
			name:     "hardcoded default go",
			db:       DatabaseTarget{defaults: &BackupDefaults{}},
			expected: "go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveBackupMethod()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveS3Prefix(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name:     "per-db override",
			db:       DatabaseTarget{S3Prefix: "production/"},
			expected: "production/",
		},
		{
			name:     "s3 config fallback",
			db:       DatabaseTarget{s3Config: &S3Config{Prefix: "backups/"}},
			expected: "backups/",
		},
		{
			name:     "per-db overrides s3 config",
			db:       DatabaseTarget{S3Prefix: "staging/", s3Config: &S3Config{Prefix: "backups/"}},
			expected: "staging/",
		},
		{
			name:     "empty when no prefix set",
			db:       DatabaseTarget{s3Config: &S3Config{}},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveS3Prefix()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveS3Bucket(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name:     "per-db override",
			db:       DatabaseTarget{S3Bucket: "my-custom-bucket"},
			expected: "my-custom-bucket",
		},
		{
			name:     "s3 config fallback",
			db:       DatabaseTarget{s3Config: &S3Config{Bucket: "default-bucket"}},
			expected: "default-bucket",
		},
		{
			name:     "per-db overrides s3 config",
			db:       DatabaseTarget{S3Bucket: "override-bucket", s3Config: &S3Config{Bucket: "default-bucket"}},
			expected: "override-bucket",
		},
		{
			name:     "empty when no bucket set",
			db:       DatabaseTarget{s3Config: &S3Config{}},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveS3Bucket()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveS3AccessKeyID(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name:     "per-db override",
			db:       DatabaseTarget{S3AccessKeyID: "custom-key-id"},
			expected: "custom-key-id",
		},
		{
			name:     "s3 config fallback",
			db:       DatabaseTarget{s3Config: &S3Config{AccessKeyID: "default-key-id"}},
			expected: "default-key-id",
		},
		{
			name:     "per-db overrides s3 config",
			db:       DatabaseTarget{S3AccessKeyID: "override-key-id", s3Config: &S3Config{AccessKeyID: "default-key-id"}},
			expected: "override-key-id",
		},
		{
			name:     "empty when no key set",
			db:       DatabaseTarget{s3Config: &S3Config{}},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveS3AccessKeyID()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestDatabaseTarget_GetEffectiveS3SecretAccessKey(t *testing.T) {
	tests := []struct {
		name     string
		db       DatabaseTarget
		expected string
	}{
		{
			name:     "per-db override",
			db:       DatabaseTarget{S3SecretAccessKey: "custom-secret"},
			expected: "custom-secret",
		},
		{
			name:     "s3 config fallback",
			db:       DatabaseTarget{s3Config: &S3Config{SecretAccessKey: "default-secret"}},
			expected: "default-secret",
		},
		{
			name:     "per-db overrides s3 config",
			db:       DatabaseTarget{S3SecretAccessKey: "override-secret", s3Config: &S3Config{SecretAccessKey: "default-secret"}},
			expected: "override-secret",
		},
		{
			name:     "empty when no secret set",
			db:       DatabaseTarget{s3Config: &S3Config{}},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.db.GetEffectiveS3SecretAccessKey()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestExpandWildcardDatabases_NonWildcard(t *testing.T) {
	target := DatabaseTarget{
		Name:     "myapp",
		Host:     "localhost",
		Port:     5432,
		Database: "myapp_db",
		User:     "admin",
		Password: "secret",
	}

	expanded, err := ExpandWildcardDatabases(target, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("expected 1 database, got %d", len(expanded))
	}

	if expanded[0].Database != "myapp_db" {
		t.Errorf("expected database 'myapp_db', got '%s'", expanded[0].Database)
	}

	if expanded[0].Name != "myapp" {
		t.Errorf("expected name 'myapp', got '%s'", expanded[0].Name)
	}
}

func TestExpandWildcardDatabases_EmptyExcludeDb(t *testing.T) {
	target := DatabaseTarget{
		Name:      "test",
		Host:      "localhost",
		Port:      5432,
		Database:  "*",
		User:      "admin",
		Password:  "secret",
		ExcludeDb: []string{},
	}

	// This test will fail if there's no PostgreSQL database running
	// So we just verify the error handling
	_, err := ExpandWildcardDatabases(target, false)
	// We expect either a connection error or successful expansion
	// If PostgreSQL is running, it should succeed; if not, it should fail gracefully
	if err != nil {
		// Check that the error is a connection error, not a logic error
		if !containsAny(err.Error(), []string{"connection", "connect", "dial", "refused"}) {
			t.Errorf("expected connection error, got: %v", err)
		}
	}
}

func TestBackupConfig_ExpandWildcards_NoWildcards(t *testing.T) {
	config := &BackupConfig{
		Databases: []DatabaseTarget{
			{
				Name:     "db1",
				Host:     "localhost",
				Port:     5432,
				Database: "app1",
				User:     "admin",
				Password: "secret",
			},
			{
				Name:     "db2",
				Host:     "localhost",
				Port:     5432,
				Database: "app2",
				User:     "admin",
				Password: "secret",
			},
		},
	}

	err := config.ExpandWildcards(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Databases) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(config.Databases))
	}

	if config.Databases[0].Database != "app1" {
		t.Errorf("expected database 'app1', got '%s'", config.Databases[0].Database)
	}

	if config.Databases[1].Database != "app2" {
		t.Errorf("expected database 'app2', got '%s'", config.Databases[1].Database)
	}
}

func TestLoadBackupConfig_WildcardDatabase(t *testing.T) {
	yaml := `
defaults:
  output_dir: "./backups"

databases:
  - name: "all_dbs"
    host: "localhost"
    port: 5432
    database: "*"
    user: "admin"
    password: "secret"
    exclude_db:
      - "test_db"
      - "temp_db"

s3:
  enabled: false
`
	path := writeTempConfig(t, yaml)
	config, err := LoadBackupConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Databases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(config.Databases))
	}
	if config.Databases[0].Database != "*" {
		t.Errorf("expected database '*', got '%s'", config.Databases[0].Database)
	}
	if len(config.Databases[0].ExcludeDb) != 2 {
		t.Errorf("expected 2 excluded databases, got %d", len(config.Databases[0].ExcludeDb))
	}
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "backup.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}
