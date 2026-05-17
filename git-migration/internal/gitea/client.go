package gitea

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const pageLimit = 50

// Owner is the repo owner as returned by the Gitea API.
type Owner struct {
	Login string `json:"login"`
}

// Repo represents a Gitea repository.
type Repo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	CloneURL    string `json:"clone_url"`
	Owner       Owner  `json:"owner"`
}

// Client is a minimal Gitea REST API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Gitea client. Trailing slashes are stripped from baseURL.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

// ListRepos returns repos for the given orgs and users.
// If both slices are empty, all repos visible to the token are returned via /repos/search.
// Repos appearing in multiple sources are deduplicated by ID.
func (c *Client) ListRepos(orgs, users []string) ([]Repo, error) {
	if len(orgs) == 0 && len(users) == 0 {
		return c.listAllRepos()
	}

	var all []Repo
	seen := make(map[int]bool)

	for _, org := range orgs {
		repos, err := c.listOrgRepos(org)
		if err != nil {
			return nil, fmt.Errorf("listing org %s: %w", org, err)
		}
		for _, r := range repos {
			if !seen[r.ID] {
				seen[r.ID] = true
				all = append(all, r)
			}
		}
	}

	for _, user := range users {
		repos, err := c.listUserRepos(user)
		if err != nil {
			return nil, fmt.Errorf("listing user %s: %w", user, err)
		}
		for _, r := range repos {
			if !seen[r.ID] {
				seen[r.ID] = true
				all = append(all, r)
			}
		}
	}

	return all, nil
}

func (c *Client) listOrgRepos(org string) ([]Repo, error) {
	return c.paginateArray(fmt.Sprintf("/api/v1/orgs/%s/repos", org))
}

func (c *Client) listUserRepos(user string) ([]Repo, error) {
	return c.paginateArray(fmt.Sprintf("/api/v1/users/%s/repos", user))
}

// listAllRepos uses /api/v1/repos/search which wraps results in {"data":[...]}.
func (c *Client) listAllRepos() ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		path := fmt.Sprintf("/api/v1/repos/search?limit=%d&page=%d", pageLimit, page)
		body, status, err := c.doGet(path)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("GET %s: status %d: %s", path, status, body)
		}

		var resp struct {
			Data []Repo `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing search response: %w", err)
		}
		all = append(all, resp.Data...)
		if len(resp.Data) < pageLimit {
			break
		}
	}
	return all, nil
}

// paginateArray pages through endpoints that return a plain JSON array of Repos.
func (c *Client) paginateArray(basePath string) ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		path := fmt.Sprintf("%s?limit=%d&page=%d", basePath, pageLimit, page)
		body, status, err := c.doGet(path)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("GET %s: status %d: %s", path, status, body)
		}

		var repos []Repo
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}
		all = append(all, repos...)
		if len(repos) < pageLimit {
			break
		}
	}
	return all, nil
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
