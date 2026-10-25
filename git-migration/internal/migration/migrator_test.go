package migration

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/adeotek/adeotek-tools/git-migration/internal/config"
	"github.com/adeotek/adeotek-tools/git-migration/internal/forgejo"
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

func newMigratorWithForgejo(t *testing.T, handler http.HandlerFunc) *Migrator {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Migrator{forgejo: forgejo.NewClient(server.URL, "tok")}
}

func TestMigrateWithFallback_disablesUnavailableFeature(t *testing.T) {
	var bodies []forgejo.MigrateRequest
	m := newMigratorWithForgejo(t, func(w http.ResponseWriter, r *http.Request) {
		var req forgejo.MigrateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		bodies = append(bodies, req)
		if req.Issues {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"message":"listing issues: not found"}`)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	err := m.migrateWithFallback(context.Background(), "org/repo", forgejo.MigrateRequest{
		RepoName: "repo",
		Issues:   true,
		Labels:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(bodies))
	}
	if bodies[1].Issues {
		t.Error("expected retry with issues disabled")
	}
	if !bodies[1].Labels {
		t.Error("expected labels to stay enabled")
	}
}

func TestMigrateWithFallback_stopsOnPersistentError(t *testing.T) {
	requests := 0
	m := newMigratorWithForgejo(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"message":"listing issues: Not Found"}`)
	})

	err := m.migrateWithFallback(context.Background(), "org/repo", forgejo.MigrateRequest{Issues: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if requests > 2 {
		t.Errorf("expected the retry loop to stop after disabling issues, got %d requests", requests)
	}
}

func TestDisabledFeature(t *testing.T) {
	cases := []struct{ msg, want string }{
		{"listing issues: not found", "issues"},
		{"listing labels: Not Found", "labels"},
		{"listing milestones: not found", "milestones"},
		{"listing pull requests: not found", "pull requests"},
		{"listing releases: not found", "releases"},
		{"listing wiki pages: not found", "wiki"},
		{"migrate returned status 500: internal error", ""},
		{"not found", ""},
	}
	for _, c := range cases {
		if got := disabledFeature(c.msg); got != c.want {
			t.Errorf("disabledFeature(%q) = %q, want %q", c.msg, got, c.want)
		}
	}
}

func TestApplyFeatureSkip_reportsChange(t *testing.T) {
	req := forgejo.MigrateRequest{Wiki: true}
	if !applyFeatureSkip(&req, "wiki") {
		t.Error("expected first skip to report a change")
	}
	if req.Wiki {
		t.Error("expected wiki to be disabled")
	}
	if applyFeatureSkip(&req, "wiki") {
		t.Error("expected second skip to report no change")
	}
	if applyFeatureSkip(&req, "unknown") {
		t.Error("expected unknown feature to report no change")
	}
}

// newGiteaRepoStub serves a single org listing with the given repos.
func newGiteaRepoStub(t *testing.T, org string, repos []gitea.Repo) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/orgs/"+org+"/repos" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("page") == "1" {
			json.NewEncoder(w).Encode(repos)
			return
		}
		json.NewEncoder(w).Encode([]gitea.Repo{})
	}))
	t.Cleanup(server.Close)
	return server
}

func TestRun_remigratePreservesPreviousRepo(t *testing.T) {
	giteaSrv := newGiteaRepoStub(t, "myorg", []gitea.Repo{{
		ID: 1, Name: "repo", FullName: "myorg/repo", Owner: gitea.Owner{Login: "myorg"}, Private: true,
	}})

	var backupName string
	var deleted, migrated bool
	forgejoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/myorg":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/myorg/repo":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/repos/myorg/repo":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			backupName = body["name"]
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/migrate":
			migrated = true
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer forgejoSrv.Close()

	m := New(gitea.NewClient(giteaSrv.URL, "tok"), forgejo.NewClient(forgejoSrv.URL, "tok"), &config.Config{
		GiteaURL:   giteaSrv.URL,
		GiteaToken: "tok",
		Orgs:       []string{"myorg"},
		OnConflict: "remigrate",
	})
	result, err := m.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Migrated != 1 {
		t.Errorf("expected 1 migrated, got %d", result.Migrated)
	}
	if !migrated {
		t.Error("expected migration to be triggered")
	}
	if !strings.HasPrefix(backupName, "repo-backup-") {
		t.Errorf("expected previous repo to be renamed to a backup, got %q", backupName)
	}
	if deleted {
		t.Error("previous repo must not be deleted without a successful migration")
	}
}

func TestRun_remigrateRestoresPreviousRepoOnFailure(t *testing.T) {
	giteaSrv := newGiteaRepoStub(t, "myorg", []gitea.Repo{{
		ID: 1, Name: "repo", FullName: "myorg/repo", Owner: gitea.Owner{Login: "myorg"},
	}})

	var backupName, restoreName string
	forgejoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/myorg":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/myorg/repo":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] == "repo" {
				restoreName = body["name"]
			} else {
				backupName = body["name"]
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/migrate":
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"message":"internal error"}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer forgejoSrv.Close()

	m := New(gitea.NewClient(giteaSrv.URL, "tok"), forgejo.NewClient(forgejoSrv.URL, "tok"), &config.Config{
		GiteaURL:   giteaSrv.URL,
		GiteaToken: "tok",
		Orgs:       []string{"myorg"},
		OnConflict: "remigrate",
	})
	result, err := m.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Failed != 1 || result.Migrated != 0 {
		t.Errorf("expected 1 failed and 0 migrated, got %+v", result)
	}
	if !strings.HasPrefix(backupName, "repo-backup-") {
		t.Errorf("expected backup rename, got %q", backupName)
	}
	if restoreName != "repo" {
		t.Errorf("expected previous repo to be restored as %q, got %q", "repo", restoreName)
	}
}

func TestRun_remigrateRestoresPreviousRepoOnCancel(t *testing.T) {
	giteaSrv := newGiteaRepoStub(t, "myorg", []gitea.Repo{{
		ID: 1, Name: "repo", FullName: "myorg/repo", Owner: gitea.Owner{Login: "myorg"},
	}})

	started := make(chan struct{})
	release := make(chan struct{})
	var mu sync.Mutex
	var backupName, restoreName string
	forgejoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/myorg":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/myorg/repo":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			mu.Lock()
			if body["name"] == "repo" {
				restoreName = body["name"]
			} else {
				backupName = body["name"]
			}
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/migrate":
			close(started)
			<-release // hold the request open until the test is done with it
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer forgejoSrv.Close()
	defer close(release) // runs before forgejoSrv.Close

	m := New(gitea.NewClient(giteaSrv.URL, "tok"), forgejo.NewClient(forgejoSrv.URL, "tok"), &config.Config{
		GiteaURL:   giteaSrv.URL,
		GiteaToken: "tok",
		Orgs:       []string{"myorg"},
		OnConflict: "remigrate",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type outcome struct {
		result Result
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := m.Run(ctx)
		done <- outcome{result, err}
	}()

	<-started
	cancel()
	got := <-done

	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", got.err)
	}
	if got.result.Failed != 1 {
		t.Errorf("expected 1 failed, got %+v", got.result)
	}
	mu.Lock()
	defer mu.Unlock()
	if !strings.HasPrefix(backupName, "repo-backup-") {
		t.Errorf("expected backup rename, got %q", backupName)
	}
	if restoreName != "repo" {
		t.Errorf("expected previous repo to be restored as %q, got %q", "repo", restoreName)
	}
}

func TestRun_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m := New(
		gitea.NewClient("http://127.0.0.1:1", "tok"),
		forgejo.NewClient("http://127.0.0.1:1", "tok"),
		&config.Config{Orgs: []string{"myorg"}},
	)
	if _, err := m.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
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
