package services

import (
	"testing"
)

// MockS3Uploader is a test mock for the S3Uploader interface
type MockS3Uploader struct {
	UploadCalls []mockUploadCall
	UploadError error
}

type mockUploadCall struct {
	LocalPath string
	Key       string
}

func (m *MockS3Uploader) Upload(localPath, key string) error {
	m.UploadCalls = append(m.UploadCalls, mockUploadCall{LocalPath: localPath, Key: key})
	return m.UploadError
}

func TestBuildS3Key(t *testing.T) {
	tests := []struct {
		name     string
		dbName   string
		filePath string
		expected string
	}{
		{
			name:     "simple path",
			dbName:   "mydb",
			filePath: "/backups/mydb_backup_20240115.sql",
			expected: "mydb/mydb_backup_20240115.sql",
		},
		{
			name:     "gzip path",
			dbName:   "mydb",
			filePath: "/backups/mydb_backup_20240115.sql.gz",
			expected: "mydb/mydb_backup_20240115.sql.gz",
		},
		{
			name:     "nested path",
			dbName:   "production",
			filePath: "/var/backups/prod/db_20240115.sql",
			expected: "production/db_20240115.sql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildS3Key(tt.dbName, tt.filePath)
			if got != tt.expected {
				t.Errorf("BuildS3Key(%q, %q) = %q, want %q", tt.dbName, tt.filePath, got, tt.expected)
			}
		})
	}
}

func TestMockS3Uploader(t *testing.T) {
	mock := &MockS3Uploader{}
	err := mock.Upload("/path/to/file.sql", "db/file.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.UploadCalls) != 1 {
		t.Fatalf("expected 1 upload call, got %d", len(mock.UploadCalls))
	}
	if mock.UploadCalls[0].LocalPath != "/path/to/file.sql" {
		t.Errorf("expected local path '/path/to/file.sql', got '%s'", mock.UploadCalls[0].LocalPath)
	}
	if mock.UploadCalls[0].Key != "db/file.sql" {
		t.Errorf("expected key 'db/file.sql', got '%s'", mock.UploadCalls[0].Key)
	}
}
