package forgejo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// requestTimeout bounds every API call so a hung server cannot block the tool forever.
const requestTimeout = 2 * time.Minute

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
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

// RepoExists returns true if the named repo exists in Forgejo.
func (c *Client) RepoExists(ctx context.Context, owner, name string) (bool, error) {
	_, status, err := c.doGet(ctx, fmt.Sprintf("/api/v1/repos/%s/%s", url.PathEscape(owner), url.PathEscape(name)))
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
// Checks for org, then user; creates an org if neither is found. A created org
// inherits the privacy of the repos being migrated.
func (c *Client) EnsureOrg(ctx context.Context, name string, private bool) error {
	escaped := url.PathEscape(name)
	_, status, err := c.doGet(ctx, "/api/v1/orgs/"+escaped)
	if err != nil {
		return err
	}
	if status == http.StatusOK {
		return nil
	}
	if status != http.StatusNotFound {
		return fmt.Errorf("checking org %q returned status %d", name, status)
	}

	_, status, err = c.doGet(ctx, "/api/v1/users/"+escaped)
	if err != nil {
		return err
	}
	if status == http.StatusOK {
		return nil
	}
	if status != http.StatusNotFound {
		return fmt.Errorf("checking user %q returned status %d", name, status)
	}

	visibility := "public"
	if private {
		visibility = "private"
	}
	body, status, err := c.doPost(ctx, "/api/v1/orgs", map[string]string{
		"username":   name,
		"visibility": visibility,
	})
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("creating org %q returned status %d: %s", name, status, strings.TrimSpace(string(body)))
	}
	return nil
}

// DeleteRepo deletes a repo from Forgejo.
func (c *Client) DeleteRepo(ctx context.Context, owner, name string) error {
	body, status, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/repos/%s/%s", url.PathEscape(owner), url.PathEscape(name)), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("delete repo %s/%s returned status %d: %s", owner, name, status, strings.TrimSpace(string(body)))
	}
	return nil
}

// RenameRepo renames a repo in Forgejo. Used to preserve the previous repo
// under a backup name before re-migrating.
func (c *Client) RenameRepo(ctx context.Context, owner, name, newName string) error {
	body, status, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/repos/%s/%s", url.PathEscape(owner), url.PathEscape(name)), map[string]string{"name": newName})
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("rename repo %s/%s to %q returned status %d: %s", owner, name, newName, status, strings.TrimSpace(string(body)))
	}
	return nil
}

// MigrateRepo triggers a server-side migration in Forgejo.
// Forgejo processes the migration asynchronously; a 201 response means it was accepted.
func (c *Client) MigrateRepo(ctx context.Context, req MigrateRequest) error {
	body, status, err := c.doPost(ctx, "/api/v1/repos/migrate", req)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("migrate returned status %d: %s", status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) doGet(ctx context.Context, path string) ([]byte, int, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

func (c *Client) doPost(ctx context.Context, path string, payload interface{}) ([]byte, int, error) {
	return c.doRequest(ctx, http.MethodPost, path, payload)
}

func (c *Client) doRequest(ctx context.Context, method, path string, payload interface{}) ([]byte, int, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}
