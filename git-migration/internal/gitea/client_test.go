package gitea

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListOrgRepos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if r.URL.Path != "/api/v1/orgs/myorg/repos" {
			http.NotFound(w, r)
			return
		}
		if page == "2" {
			json.NewEncoder(w).Encode([]Repo{})
			return
		}
		json.NewEncoder(w).Encode([]Repo{
			{ID: 1, Name: "repo1", FullName: "myorg/repo1", Owner: Owner{Login: "myorg"}},
			{ID: 2, Name: "repo2", FullName: "myorg/repo2", Owner: Owner{Login: "myorg"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	repos, err := c.listOrgRepos("myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].Name != "repo1" {
		t.Errorf("expected repo1, got %s", repos[0].Name)
	}
}

func TestListOrgRepos_setsAuthHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode([]Repo{})
	}))
	defer server.Close()

	c := NewClient(server.URL, "my-secret-token")
	c.listOrgRepos("org")
	if gotAuth != "token my-secret-token" {
		t.Errorf("expected 'token my-secret-token', got %q", gotAuth)
	}
}

func TestListAllRepos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "2" {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": []Repo{}, "ok": true})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []Repo{{ID: 1, Name: "repo1", FullName: "org/repo1"}},
			"ok":   true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	repos, err := c.listAllRepos()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(repos))
	}
}

func TestListRepos_deduplicates(t *testing.T) {
	// Both org and user endpoints return the same repo ID — should appear once.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Repo{
			{ID: 1, Name: "shared", FullName: "org/shared", Owner: Owner{Login: "org"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	repos, err := c.ListRepos([]string{"org"}, []string{"user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo after dedup, got %d", len(repos))
	}
}

func TestListRepos_noFilter_callsSearch(t *testing.T) {
	var searchCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/repos/search" {
			searchCalled = true
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []Repo{}, "ok": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok")
	c.ListRepos(nil, nil)
	if !searchCalled {
		t.Error("expected /api/v1/repos/search to be called when no orgs/users given")
	}
}
