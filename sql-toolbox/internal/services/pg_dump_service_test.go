package services

import (
	"testing"
	"time"
)

func TestEscapeStringLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no escaping needed",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "single quote",
			input:    "it's a test",
			expected: "it''s a test",
		},
		{
			name:     "backslash",
			input:    `path\to\file`,
			expected: `path\\to\\file`,
		},
		{
			name:     "both quotes and backslashes",
			input:    `it's a \test`,
			expected: `it''s a \\test`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple single quotes",
			input:    "a''b",
			expected: "a''''b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeStringLiteral(tt.input)
			if got != tt.expected {
				t.Errorf("escapeStringLiteral(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatInsertValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		colType  string
		expected string
	}{
		{
			name:     "nil value",
			value:    nil,
			colType:  "text",
			expected: "NULL",
		},
		{
			name:     "bool true",
			value:    true,
			colType:  "boolean",
			expected: "true",
		},
		{
			name:     "bool false",
			value:    false,
			colType:  "boolean",
			expected: "false",
		},
		{
			name:     "integer",
			value:    int64(42),
			colType:  "integer",
			expected: "42",
		},
		{
			name:     "negative integer",
			value:    int64(-10),
			colType:  "bigint",
			expected: "-10",
		},
		{
			name:     "float",
			value:    3.14,
			colType:  "double precision",
			expected: "3.14",
		},
		{
			name:     "simple string",
			value:    "hello",
			colType:  "text",
			expected: "'hello'",
		},
		{
			name:     "string with quotes",
			value:    "it's",
			colType:  "text",
			expected: "'it''s'",
		},
		{
			name:     "bytea column",
			value:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
			colType:  "bytea",
			expected: "'\\xdeadbeef'",
		},
		{
			name:     "bytes as text",
			value:    []byte("hello"),
			colType:  "text",
			expected: "'hello'",
		},
		{
			name:     "timestamp",
			value:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			colType:  "timestamp with time zone",
			expected: "'2024-01-15 10:30:00+00'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatInsertValue(tt.value, tt.colType)
			if got != tt.expected {
				t.Errorf("formatInsertValue(%v, %q) = %q, want %q", tt.value, tt.colType, got, tt.expected)
			}
		})
	}
}

func TestParsePostgresArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty",
			input:    "{}",
			expected: nil,
		},
		{
			name:     "single element",
			input:    "{hello}",
			expected: []string{"hello"},
		},
		{
			name:     "multiple elements",
			input:    "{a,b,c}",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "quoted elements",
			input:    `{"hello world","foo bar"}`,
			expected: []string{"hello world", "foo bar"},
		},
		{
			name:     "escaped quote",
			input:    `{"it\"s"}`,
			expected: []string{`it"s`},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePostgresArray(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("parsePostgresArray(%q) = %v, want %v", tt.input, got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("parsePostgresArray(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", `"simple"`},
		{"my_table", `"my_table"`},
		{"public", `"public"`},
	}
	for _, tt := range tests {
		got := quoteIdentifier(tt.input)
		if got != tt.expected {
			t.Errorf("quoteIdentifier(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestQualifiedName(t *testing.T) {
	got := qualifiedName("public", "users")
	expected := `"public"."users"`
	if got != expected {
		t.Errorf("qualifiedName(\"public\", \"users\") = %q, want %q", got, expected)
	}
}
