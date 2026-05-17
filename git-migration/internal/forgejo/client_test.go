package forgejo

import (
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
	exists, err := c.RepoExists("owner", "repo")
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
	exists, err := c.RepoExists("owner", "repo")
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
	if err := c.EnsureOrg("myorg"); err != nil {
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
	if err := c.EnsureOrg("someuser"); err != nil {
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
	if err := c.EnsureOrg("neworg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected org to be created")
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
	if err := c.MigrateRepo(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.RepoName != "myrepo" {
		t.Errorf("expected myrepo, got %s", gotBody.RepoName)
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
	c.MigrateRepo(MigrateRequest{})
	if gotAuth != "token my-forgejo-token" {
		t.Errorf("expected 'token my-forgejo-token', got %q", gotAuth)
	}
}
