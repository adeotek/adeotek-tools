package models

import (
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
			name: "Raw connection string",
			params: ConnectionParameters{
				Provider:            "postgresql",
				RawConnectionString: "custom connection string",
			},
			expected: "custom connection string",
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
