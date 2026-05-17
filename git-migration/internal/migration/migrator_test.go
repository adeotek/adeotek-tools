package migration

import (
	"testing"

	"github.com/adeotek/adeotek-tools/git-migration/internal/config"
	"github.com/adeotek/adeotek-tools/git-migration/internal/gitea"
)

func TestFilter_glob_match(t *testing.T) {
	m := &Migrator{cfg: &config.Config{Filter: "infra-*"}}
	repos := []gitea.Repo{
		{Name: "infra-docker", FullName: "org/infra-docker"},
		{Name: "infra-k8s", FullName: "org/infra-k8s"},
		{Name: "app-web", FullName: "org/app-web"},
	}
	got := m.filter(repos)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(got), got)
	}
	if got[0].Name != "infra-docker" || got[1].Name != "infra-k8s" {
		t.Errorf("unexpected repos: %v", got)
	}
}

func TestFilter_glob_no_match(t *testing.T) {
	m := &Migrator{cfg: &config.Config{Filter: "xyz-*"}}
	repos := []gitea.Repo{
		{Name: "infra-docker", FullName: "org/infra-docker"},
	}
	got := m.filter(repos)
	if len(got) != 0 {
		t.Errorf("expected 0, got %d", len(got))
	}
}

func TestFilter_exclude_by_name(t *testing.T) {
	m := &Migrator{cfg: &config.Config{Exclude: []string{"secret-repo"}}}
	repos := []gitea.Repo{
		{Name: "public-repo", FullName: "org/public-repo"},
		{Name: "secret-repo", FullName: "org/secret-repo"},
	}
	got := m.filter(repos)
	if len(got) != 1 || got[0].Name != "public-repo" {
		t.Errorf("unexpected: %v", got)
	}
}

func TestFilter_exclude_by_full_name(t *testing.T) {
	m := &Migrator{cfg: &config.Config{Exclude: []string{"org/secret-repo"}}}
	repos := []gitea.Repo{
		{Name: "public-repo", FullName: "org/public-repo"},
		{Name: "secret-repo", FullName: "org/secret-repo"},
	}
	got := m.filter(repos)
	if len(got) != 1 || got[0].Name != "public-repo" {
		t.Errorf("unexpected: %v", got)
	}
}

func TestFilter_no_filter_returns_all(t *testing.T) {
	m := &Migrator{cfg: &config.Config{}}
	repos := []gitea.Repo{{Name: "a"}, {Name: "b"}}
	got := m.filter(repos)
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}

func TestDestOwner_mapped(t *testing.T) {
	m := &Migrator{cfg: &config.Config{
		OrgMappings: config.OrgMappings{"gitea-org": "forgejo-org"},
	}}
	if got := m.destOwner("gitea-org"); got != "forgejo-org" {
		t.Errorf("expected forgejo-org, got %s", got)
	}
}

func TestDestOwner_unmapped_returns_original(t *testing.T) {
	m := &Migrator{cfg: &config.Config{OrgMappings: config.OrgMappings{}}}
	if got := m.destOwner("some-org"); got != "some-org" {
		t.Errorf("expected some-org, got %s", got)
	}
}
