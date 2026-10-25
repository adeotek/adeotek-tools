package forgejo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepoExists_true(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	exists, err := c.RepoExists(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
}

func TestRepoExists_false(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	exists, err := c.RepoExists(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestEnsureOrg_orgAlreadyExists(t *testing.T) {
	var createCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/orgs/myorg" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/v1/orgs" && r.Method == http.MethodPost {
			createCalled = true
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.EnsureOrg(context.Background(), "myorg", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createCalled {
		t.Error("org already existed; create should not have been called")
	}
}

func TestEnsureOrg_userExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/orgs/someuser" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Path == "/api/v1/users/someuser" {
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.EnsureOrg(context.Background(), "someuser", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureOrg_createsWhenMissing(t *testing.T) {
	var created bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/orgs/neworg" && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.URL.Path == "/api/v1/users/neworg" && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.URL.Path == "/api/v1/orgs" && r.Method == http.MethodPost:
			created = true
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.EnsureOrg(context.Background(), "neworg", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected org to be created")
	}
}

func TestEnsureOrg_escapesPath(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.EnsureOrg(context.Background(), "my org", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "/api/v1/orgs/my%20org"; len(paths) != 1 || paths[0] != want {
		t.Errorf("expected path %q, got %v", want, paths)
	}
}

func TestEnsureOrg_createsPrivateOrgForPrivateRepo(t *testing.T) {
	var visibility string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/orgs":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			visibility = body["visibility"]
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.EnsureOrg(context.Background(), "neworg", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if visibility != "private" {
		t.Errorf("expected private org visibility, got %q", visibility)
	}
}

func TestEnsureOrg_createsPublicOrgForPublicRepo(t *testing.T) {
	var visibility string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			visibility = body["visibility"]
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.EnsureOrg(context.Background(), "neworg", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if visibility != "public" {
		t.Errorf("expected public org visibility, got %q", visibility)
	}
}

func TestRenameRepo_success(t *testing.T) {
	var gotPath, gotName string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		gotName = body["name"]
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.RenameRepo(context.Background(), "owner", "repo", "repo-backup-20260101-000000"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/repos/owner/repo" {
		t.Errorf("unexpected path %q", gotPath)
	}
	if gotName != "repo-backup-20260101-000000" {
		t.Errorf("unexpected new name %q", gotName)
	}
}

func TestRenameRepo_failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.RenameRepo(context.Background(), "owner", "repo", "taken"); err == nil {
		t.Error("expected error for non-200 rename response")
	}
}

func TestMigrateRepo_success(t *testing.T) {
	var gotBody MigrateRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	req := MigrateRequest{
		RepoName:  "myrepo",
		RepoOwner: "myorg",
		CloneAddr: "https://gitea.lan/myorg/myrepo",
		Service:   "gitea",
		Issues:    true,
	}
	if err := c.MigrateRepo(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.RepoName != "myrepo" {
		t.Errorf("expected myrepo, got %s", gotBody.RepoName)
	}
}

func TestDeleteRepo_success(t *testing.T) {
	var deleted bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/owner/repo" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	if err := c.DeleteRepo(context.Background(), "owner", "repo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
}

func TestMigrateRepo_setsAuthHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := NewClient(server.URL, "my-forgejo-token")
	c.MigrateRepo(context.Background(), MigrateRequest{})
	if gotAuth != "token my-forgejo-token" {
		t.Errorf("expected 'token my-forgejo-token', got %q", gotAuth)
	}
}
