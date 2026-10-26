// Package filter provides repository filtering functionality
package filter

import (
	"fmt"
	"path"
	"strings"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
)

// FilterRepositories applies include and exclude filters from config.
//
// Filter entries are repository full names (e.g. "owner/repo") and may contain
// the `*` wildcard (e.g. "owner/*", "owner/prefix-*"). A `*` matches any
// sequence of characters except the path separator, so "owner/*" matches all
// repositories of an owner but not repositories of a nested namespace.
func FilterRepositories(repos []repository.Repository, provider *config.ProviderConfig, verbose bool) []repository.Repository {
	// If neither include nor exclude is specified, return all repos
	if len(provider.Include) == 0 && len(provider.Exclude) == 0 {
		return repos
	}

	if verbose {
		if len(provider.Include) == 0 {
			fmt.Printf("----> `include` filter is empty\n")
		} else {
			fmt.Printf("----> `include` filter: %s\n", strings.Join(provider.Include, ", "))
		}
		if len(provider.Exclude) == 0 {
			fmt.Printf("----> `exclude` filter is empty\n")
		} else {
			fmt.Printf("----> `exclude` filter: %s\n", strings.Join(provider.Exclude, ", "))
		}
	}

	var filtered []repository.Repository
	for _, repo := range repos {
		// If include list is specified, only include repos matching it
		if len(provider.Include) > 0 {
			if matchesAny(repo.FullName, provider.Include) {
				filtered = append(filtered, repo)
			}
		} else {
			// If exclude list is specified, exclude repos matching it
			if !matchesAny(repo.FullName, provider.Exclude) {
				filtered = append(filtered, repo)
			}
		}
	}

	return filtered
}

// matchesAny reports whether name matches at least one of the given patterns.
func matchesAny(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern reports whether name matches pattern. Patterns without a
// wildcard are compared literally; patterns containing `*` are matched using
// shell-style matching, where `*` does not cross the `/` separator.
func matchesPattern(name, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return name == pattern
	}

	matched, err := path.Match(pattern, name)
	if err != nil {
		// Treat an invalid pattern as a literal repository name.
		return name == pattern
	}
	return matched
}
