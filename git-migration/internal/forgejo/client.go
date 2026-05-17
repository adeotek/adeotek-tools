package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MigrateRequest is the payload for POST /api/v1/repos/migrate.
type MigrateRequest struct {
	AuthToken    string `json:"auth_token"`
	CloneAddr    string `json:"clone_addr"`
	Description  string `json:"description"`
	Issues       bool   `json:"issues"`
	Labels       bool   `json:"labels"`
	Milestones   bool   `json:"milestones"`
	Mirror       bool   `json:"mirror"`
	Private      bool   `json:"private"`
	PullRequests bool   `json:"pull_requests"`
	Releases     bool   `json:"releases"`
	RepoName     string `json:"repo_name"`
	RepoOwner    string `json:"repo_owner"`
	Wiki         bool   `json:"wiki"`
	Service      string `json:"service"`
}

// Client is a minimal Forgejo REST API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Forgejo client. Trailing slashes are stripped from baseURL.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

// RepoExists returns true if the named repo exists in Forgejo.
func (c *Client) RepoExists(owner, name string) (bool, error) {
	_, status, err := c.doGet(fmt.Sprintf("/api/v1/repos/%s/%s", owner, name))
	if err != nil {
		return false, err
	}
	switch status {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status %d checking repo %s/%s", status, owner, name)
	}
}

// EnsureOrg ensures the named owner exists in Forgejo.
// Checks for org, then user; creates an org if neither is found.
func (c *Client) EnsureOrg(name string) error {
	_, status, err := c.doGet("/api/v1/orgs/" + name)
	if err != nil {
		return err
	}
	if status == http.StatusOK {
		return nil
	}
	if status != http.StatusNotFound {
		return fmt.Errorf("checking org %q returned status %d", name, status)
	}

	_, status, err = c.doGet("/api/v1/users/" + name)
	if err != nil {
		return err
	}
	if status == http.StatusOK {
		return nil
	}
	if status != http.StatusNotFound {
		return fmt.Errorf("checking user %q returned status %d", name, status)
	}

	body, status, err := c.doPost("/api/v1/orgs", map[string]string{
		"username":   name,
		"visibility": "public",
	})
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("creating org %q returned status %d: %s", name, status, strings.TrimSpace(string(body)))
	}
	return nil
}

// MigrateRepo triggers a server-side migration in Forgejo.
// Forgejo processes the migration asynchronously; a 201 response means it was accepted.
func (c *Client) MigrateRepo(req MigrateRequest) error {
	body, status, err := c.doPost("/api/v1/repos/migrate", req)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("migrate returned status %d: %s", status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) doGet(path string) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (c *Client) doPost(path string, payload interface{}) ([]byte, int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}
