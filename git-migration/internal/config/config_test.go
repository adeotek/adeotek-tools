package config

import (
	"os"
	"testing"
)

func TestOrgMappings_Set_valid(t *testing.T) {
	m := make(OrgMappings)
	if err := m.Set("src:dst"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["src"] != "dst" {
		t.Errorf("expected dst, got %s", m["src"])
	}
}

func TestOrgMappings_Set_invalid(t *testing.T) {
	cases := []string{"nodash", ":empty-src", "empty-dst:", ""}
	for _, c := range cases {
		m := make(OrgMappings)
		if err := m.Set(c); err == nil {
			t.Errorf("expected error for %q, got nil", c)
		}
	}
}

func TestOrgMappings_String(t *testing.T) {
	m := OrgMappings{"a": "b"}
	if m.String() == "" {
		t.Error("expected non-empty string")
	}
}

func TestSplitList(t *testing.T) {
	got := SplitList("a, b , c")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestSplitList_empty(t *testing.T) {
	if SplitList("") != nil {
		t.Error("expected nil for empty string")
	}
}

func TestResolveToken_flag_takes_priority(t *testing.T) {
	os.Setenv("TEST_TOKEN", "env-val")
	defer os.Unsetenv("TEST_TOKEN")
	if got := ResolveToken("flag-val", "TEST_TOKEN"); got != "flag-val" {
		t.Errorf("expected flag-val, got %s", got)
	}
}

func TestResolveToken_falls_back_to_env(t *testing.T) {
	os.Setenv("TEST_TOKEN", "env-val")
	defer os.Unsetenv("TEST_TOKEN")
	if got := ResolveToken("", "TEST_TOKEN"); got != "env-val" {
		t.Errorf("expected env-val, got %s", got)
	}
}

func TestValidate_missing_fields(t *testing.T) {
	cfg := &Config{OnConflict: "skip"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing gitea-url")
	}
	cfg.GiteaURL = "http://gitea.lan"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing gitea-token")
	}
	cfg.GiteaToken = "tok"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing forgejo-url")
	}
	cfg.ForgejoURL = "http://forgejo.lan"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing forgejo-token")
	}
	cfg.ForgejoToken = "tok"
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_invalid_on_conflict(t *testing.T) {
	cfg := &Config{
		GiteaURL: "http://gitea.lan", GiteaToken: "tok",
		ForgejoURL: "http://forgejo.lan", ForgejoToken: "tok",
		OnConflict: "invalid",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid on-conflict value")
	}
}

func TestValidate_invalid_filter(t *testing.T) {
	cfg := &Config{
		GiteaURL: "http://gitea.lan", GiteaToken: "tok",
		ForgejoURL: "http://forgejo.lan", ForgejoToken: "tok",
		OnConflict: "skip",
		Filter:     "[bad",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid glob pattern")
	}
}
