package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	pageLimit      = 50
	requestTimeout = 2 * time.Minute
)

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
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

// ListRepos returns repos for the given orgs and users.
// If both slices are empty, all repos visible to the token are returned via /repos/search.
// Repos appearing in multiple sources are deduplicated by ID.
func (c *Client) ListRepos(ctx context.Context, orgs, users []string) ([]Repo, error) {
	if len(orgs) == 0 && len(users) == 0 {
		return c.listAllRepos(ctx)
	}

	var all []Repo
	seen := make(map[int]bool)

	for _, org := range orgs {
		repos, err := c.listOrgRepos(ctx, org)
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
		repos, err := c.listUserRepos(ctx, user)
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

func (c *Client) listOrgRepos(ctx context.Context, org string) ([]Repo, error) {
	return c.paginateArray(ctx, fmt.Sprintf("/api/v1/orgs/%s/repos", url.PathEscape(org)))
}

func (c *Client) listUserRepos(ctx context.Context, user string) ([]Repo, error) {
	return c.paginateArray(ctx, fmt.Sprintf("/api/v1/users/%s/repos", url.PathEscape(user)))
}

// listAllRepos uses /api/v1/repos/search which wraps results in {"data":[...]}.
func (c *Client) listAllRepos(ctx context.Context) ([]Repo, error) {
	return c.paginate(ctx, func(page int) ([]Repo, error) {
		path := fmt.Sprintf("/api/v1/repos/search?limit=%d&page=%d", pageLimit, page)
		body, status, err := c.doGet(ctx, path)
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
		return resp.Data, nil
	})
}

// paginateArray pages through endpoints that return a plain JSON array of Repos.
func (c *Client) paginateArray(ctx context.Context, basePath string) ([]Repo, error) {
	return c.paginate(ctx, func(page int) ([]Repo, error) {
		path := fmt.Sprintf("%s?limit=%d&page=%d", basePath, pageLimit, page)
		body, status, err := c.doGet(ctx, path)
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
		return repos, nil
	})
}

// paginate walks pages until a page yields no new repos. Servers may cap the
// page size below pageLimit, so a short page is not treated as the last one;
// stopping on repeated IDs also guards against servers that ignore the page
// parameter.
func (c *Client) paginate(ctx context.Context, fetch func(page int) ([]Repo, error)) ([]Repo, error) {
	var all []Repo
	seen := make(map[int]bool)
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		repos, err := fetch(page)
		if err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		added := 0
		for _, r := range repos {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			all = append(all, r)
			added++
		}
		if added == 0 {
			break
		}
	}
	return all, nil
}

func (c *Client) doGet(ctx context.Context, path string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
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
