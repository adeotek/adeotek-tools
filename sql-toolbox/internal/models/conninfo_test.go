package models

import (
	"strings"
	"testing"
)

func TestParseConnectionParameters_DefaultPostgresPort(t *testing.T) {
	params, err := ParseConnectionParameters("postgresql", "", "localhost", "", "mydb", "user", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Port != 5432 {
		t.Errorf("expected default port 5432, got %d", params.Port)
	}
	if valid, errs := params.IsValid(); !valid {
		t.Errorf("expected valid parameters, got %v", errs)
	}
}

func TestGetConnectionString_QuotesSpecialValues(t *testing.T) {
	params := &ConnectionParameters{
		Provider:     "postgresql",
		Host:         "localhost",
		Port:         5432,
		DatabaseName: "my db",
		User:         "user",
		Password:     `pa ss'wo\rd`,
	}
	connStr, err := params.GetConnectionString()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(connStr, `dbname='my db'`) {
		t.Errorf("expected quoted dbname, got %q", connStr)
	}
	if !strings.Contains(connStr, `password='pa ss\'wo\\rd'`) {
		t.Errorf("expected escaped password, got %q", connStr)
	}

	clean, password := RemovePasswordFromConnInfo(connStr)
	if password != `pa ss'wo\rd` {
		t.Errorf("expected password to round-trip, got %q", password)
	}
	if strings.Contains(clean, "password") {
		t.Errorf("expected password to be removed, got %q", clean)
	}
	if !strings.Contains(clean, `dbname='my db'`) {
		t.Errorf("expected other values to survive, got %q", clean)
	}
}

func TestRewriteConnInfoHostPort(t *testing.T) {
	got := RewriteConnInfoHostPort("host=db.internal port=5432 dbname=mydb password=secret", "127.0.0.1", 40000)
	if !strings.Contains(got, "host=127.0.0.1") || !strings.Contains(got, "port=40000") {
		t.Errorf("expected host/port rewrite, got %q", got)
	}
	if !strings.Contains(got, "dbname=mydb") || !strings.Contains(got, "password=secret") {
		t.Errorf("expected other parameters to survive, got %q", got)
	}
}

func TestConvertDotNetFormat_QuotedSemicolon(t *testing.T) {
	got, err := normalizeConnectionString(`Server=localhost;Port=5432;Database=mydb;Username=u;Password="pa;ss";`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, `password=pa;ss`) {
		t.Errorf("expected password with semicolon to be preserved, got %q", got)
	}
}
