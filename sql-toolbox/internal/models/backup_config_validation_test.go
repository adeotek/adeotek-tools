package models

import "testing"

func TestBackupConfig_Validate_RejectsUnknownBackupMethod(t *testing.T) {
	method := "pgdump"
	config := BackupConfig{
		Defaults: BackupDefaults{BackupMethod: &method},
		Databases: []DatabaseTarget{
			{Name: "db", Host: "localhost", Database: "db"},
		},
	}
	if err := config.Validate(); err == nil {
		t.Fatal("expected error for unknown backup_method")
	}
}

func TestBackupConfig_Validate_UploadRequiresS3Enabled(t *testing.T) {
	config := BackupConfig{
		Defaults: BackupDefaults{UploadToS3: boolPtr(true)},
		Databases: []DatabaseTarget{
			{Name: "db", Host: "localhost", Database: "db"},
		},
	}
	config.Databases[0].SetDefaults(&config.Defaults, &config.S3)
	if err := config.Validate(); err == nil {
		t.Fatal("expected error when upload_to_s3 is enabled but s3.enabled is false")
	}
}

func TestBackupConfig_Validate_RejectsPartialS3Credentials(t *testing.T) {
	config := BackupConfig{
		Databases: []DatabaseTarget{
			{Name: "db", Host: "localhost", Database: "db"},
		},
		S3: S3Config{Enabled: true, Bucket: "bucket", AccessKeyID: "key-only"},
	}
	if err := config.Validate(); err == nil {
		t.Fatal("expected error for partial S3 credentials")
	}
}
