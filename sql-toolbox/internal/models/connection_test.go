package models

import (
	"strings"
	"testing"
)

func TestParseDatabaseProvider(t *testing.T) {
	tests := []struct {
		input    string
		expected DatabaseProvider
	}{
		{"postgresql", PostgreSQL},
		{"postgres", PostgreSQL},
		{"PostgreSQL", PostgreSQL},
		{"sqlite", SQLite},
		{"SQLite", SQLite},
		{"unknown", Unknown},
		{"", Unknown},
	}

	for _, test := range tests {
		result := ParseDatabaseProvider(test.input)
		if result != test.expected {
			t.Errorf("ParseDatabaseProvider(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestConnectionParameters_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		params   ConnectionParameters
		expected bool
	}{
		{
			name: "Valid PostgreSQL with individual parameters",
			params: ConnectionParameters{
				Provider:     "postgresql",
				Host:         "localhost",
				Port:         5432,
				DatabaseName: "testdb",
				User:         "testuser",
				Password:     "testpass",
			},
			expected: true,
		},
		{
			name: "Valid SQLite",
			params: ConnectionParameters{
				Provider:     "sqlite",
				DatabaseName: "test.db",
			},
			expected: true,
		},
		{
			name: "Valid with connection string",
			params: ConnectionParameters{
				Provider:            "postgresql",
				RawConnectionString: "host=localhost port=5432 dbname=testdb user=testuser password=testpass",
			},
			expected: true,
		},
		{
			name: "Invalid provider",
			params: ConnectionParameters{
				Provider: "unknown",
			},
			expected: false,
		},
		{
			name: "Missing database name",
			params: ConnectionParameters{
				Provider: "postgresql",
				Host:     "localhost",
				Port:     5432,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, _ := test.params.IsValid()
			if valid != test.expected {
				t.Errorf("IsValid() = %v, expected %v", valid, test.expected)
			}
		})
	}
}

func TestConnectionParameters_GetConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		params   ConnectionParameters
		expected string
	}{
		{
			name: "PostgreSQL connection string",
			params: ConnectionParameters{
				Provider:     "postgresql",
				Host:         "localhost",
				Port:         5432,
				DatabaseName: "testdb",
				User:         "testuser",
				Password:     "testpass",
			},
			expected: "host=localhost port=5432 dbname=testdb user=testuser password=testpass sslmode=disable",
		},
		{
			name: "SQLite connection string",
			params: ConnectionParameters{
				Provider:     "sqlite",
				DatabaseName: "test.db",
			},
			expected: "test.db",
		},
		{
			name: "Raw connection string (lib/pq format)",
			params: ConnectionParameters{
				Provider:            "postgresql",
				RawConnectionString: "host=myhost port=5432 dbname=mydb user=dbuser password=dbpass sslmode=disable",
			},
			expected: "host=myhost port=5432 dbname=mydb user=dbuser password=dbpass sslmode=disable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.params.GetConnectionString()
			if err != nil {
				t.Errorf("GetConnectionString() error = %v", err)
				return
			}
			if result != test.expected {
				t.Errorf("GetConnectionString() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestNormalizeConnectionString_LibPQFormat(t *testing.T) {
	// Test that lib/pq format passes through unchanged
	input := "host=localhost port=5432 dbname=mydb user=dbuser password=dbpass sslmode=disable"
	result, err := normalizeConnectionString(input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != input {
		t.Errorf("Expected unchanged string, got: %s", result)
	}
}

func TestNormalizeConnectionString_DotNetFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string // key-value pairs that should be present
	}{
		{
			name:  "Standard .NET format",
			input: "Server=myhost;Port=5432;Database=mydb;Username=dbuser;Password=dbpassword;SslMode=Disable;",
			expected: map[string]string{
				"host":     "myhost",
				"port":     "5432",
				"dbname":   "mydb",
				"user":     "dbuser",
				"password": "dbpassword",
				"sslmode":  "disable",
			},
		},
		{
			name:  ".NET format without trailing semicolon",
			input: "Server=localhost;Port=5433;Database=testdb;Username=admin;Password=secret",
			expected: map[string]string{
				"host":     "localhost",
				"port":     "5433",
				"dbname":   "testdb",
				"user":     "admin",
				"password": "secret",
			},
		},
		{
			name:  ".NET format with mixed case",
			input: "SERVER=myhost;PORT=5432;DATABASE=mydb;USERNAME=dbuser;PASSWORD=dbpassword",
			expected: map[string]string{
				"host":     "myhost",
				"port":     "5432",
				"dbname":   "mydb",
				"user":     "dbuser",
				"password": "dbpassword",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeConnectionString(tt.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			// Parse result into key-value pairs
			pairs := strings.Split(result, " ")
			resultMap := make(map[string]string)
			for _, pair := range pairs {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					resultMap[parts[0]] = parts[1]
				}
			}

			// Check that all expected key-value pairs are present
			for key, expectedValue := range tt.expected {
				if actualValue, ok := resultMap[key]; !ok {
					t.Errorf("Expected key '%s' not found in result", key)
				} else if actualValue != expectedValue {
					t.Errorf("For key '%s', expected value '%s', got '%s'", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestNormalizeConnectionString_PostgreSQLURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "Standard PostgreSQL URL",
			input: "postgresql://dbuser:dbpassword@myhost:5432/mydb",
			expected: map[string]string{
				"host":     "myhost",
				"port":     "5432",
				"dbname":   "mydb",
				"user":     "dbuser",
				"password": "dbpassword",
			},
		},
		{
			name:  "PostgreSQL URL with postgres:// scheme",
			input: "postgres://admin:secret@localhost:5433/testdb",
			expected: map[string]string{
				"host":     "localhost",
				"port":     "5433",
				"dbname":   "testdb",
				"user":     "admin",
				"password": "secret",
			},
		},
		{
			name:  "PostgreSQL URL without port (should default to 5432)",
			input: "postgresql://dbuser:dbpass@myhost/mydb",
			expected: map[string]string{
				"host":     "myhost",
				"port":     "5432",
				"dbname":   "mydb",
				"user":     "dbuser",
				"password": "dbpass",
			},
		},
		{
			name:  "PostgreSQL URL with sslmode parameter",
			input: "postgresql://dbuser:dbpass@myhost:5432/mydb?sslmode=require",
			expected: map[string]string{
				"host":     "myhost",
				"port":     "5432",
				"dbname":   "mydb",
				"user":     "dbuser",
				"password": "dbpass",
				"sslmode":  "require",
			},
		},
		{
			name:  "PostgreSQL URL without password",
			input: "postgresql://dbuser@myhost:5432/mydb",
			expected: map[string]string{
				"host":   "myhost",
				"port":   "5432",
				"dbname": "mydb",
				"user":   "dbuser",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeConnectionString(tt.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			// Parse result into key-value pairs
			pairs := strings.Split(result, " ")
			resultMap := make(map[string]string)
			for _, pair := range pairs {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					resultMap[parts[0]] = parts[1]
				}
			}

			// Check that all expected key-value pairs are present
			for key, expectedValue := range tt.expected {
				if actualValue, ok := resultMap[key]; !ok {
					t.Errorf("Expected key '%s' not found in result", key)
				} else if actualValue != expectedValue {
					t.Errorf("For key '%s', expected value '%s', got '%s'", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestGetConnectionString_WithNormalization(t *testing.T) {
	tests := []struct {
		name             string
		rawConnString    string
		expectedContains []string
		shouldError      bool
	}{
		{
			name:          ".NET format connection string",
			rawConnString: "Server=myhost;Port=5432;Database=mydb;Username=dbuser;Password=dbpass",
			expectedContains: []string{
				"host=myhost",
				"port=5432",
				"dbname=mydb",
				"user=dbuser",
				"password=dbpass",
			},
			shouldError: false,
		},
		{
			name:          "PostgreSQL URL format",
			rawConnString: "postgresql://dbuser:dbpass@myhost:5432/mydb",
			expectedContains: []string{
				"host=myhost",
				"port=5432",
				"dbname=mydb",
				"user=dbuser",
				"password=dbpass",
			},
			shouldError: false,
		},
		{
			name:          "lib/pq format (unchanged)",
			rawConnString: "host=myhost port=5432 dbname=mydb user=dbuser password=dbpass",
			expectedContains: []string{
				"host=myhost",
				"port=5432",
				"dbname=mydb",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &ConnectionParameters{
				Provider:            "postgresql",
				RawConnectionString: tt.rawConnString,
			}

			result, err := cp.GetConnectionString()

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}
