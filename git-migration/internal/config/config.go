package config

import (
	"fmt"
	"os"
	"strings"
)

// OrgMappings maps source org names to destination org names.
// Implements flag.Value so it can be used as a repeatable --map-org flag.
type OrgMappings map[string]string

func (m OrgMappings) String() string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+":"+v)
	}
	return strings.Join(parts, ",")
}

func (m OrgMappings) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid mapping %q: expected format src:dst", value)
	}
	m[parts[0]] = parts[1]
	return nil
}

// Config holds all runtime configuration for git-migration.
type Config struct {
	GiteaURL     string
	GiteaToken   string
	ForgejoURL   string
	ForgejoToken string
	Orgs         []string
	Users        []string
	Filter       string
	Exclude      []string
	OrgMappings  OrgMappings
	OnConflict   string
	DryRun       bool
	Verbose      bool
}

// ResolveToken returns flagVal if non-empty, otherwise the env var named by envKey.
func ResolveToken(flagVal, envKey string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv(envKey)
}

// SplitList splits a comma-separated string into a trimmed, non-empty slice.
func SplitList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

// Validate returns an error if required fields are missing or invalid.
func (c *Config) Validate() error {
	if c.GiteaURL == "" {
		return fmt.Errorf("--gitea-url is required")
	}
	if c.GiteaToken == "" {
		return fmt.Errorf("--gitea-token or GITEA_TOKEN is required")
	}
	if c.ForgejoURL == "" {
		return fmt.Errorf("--forgejo-url is required")
	}
	if c.ForgejoToken == "" {
		return fmt.Errorf("--forgejo-token or FORGEJO_TOKEN is required")
	}
	switch c.OnConflict {
	case "skip", "fail", "remigrate":
	default:
		return fmt.Errorf("--on-conflict must be skip, fail, or remigrate (got %q)", c.OnConflict)
	}
	return nil
}
