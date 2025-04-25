// Package filter provides repository filtering functionality
package filter

import (
	"github.com/adeotek/git-multi-repo-clone/internal/config"
	"github.com/adeotek/git-multi-repo-clone/internal/repository"
)

// FilterRepositories applies include and exclude filters from config
func FilterRepositories(repos []repository.Repository, config *config.Config) []repository.Repository {
	// If neither include nor exclude is specified, return all repos
	if len(config.Include) == 0 && len(config.Exclude) == 0 {
		return repos
	}

	// Create a map for faster lookup
	includeMap := make(map[string]bool)
	for _, name := range config.Include {
		includeMap[name] = true
	}

	excludeMap := make(map[string]bool)
	for _, name := range config.Exclude {
		excludeMap[name] = true
	}

	var filtered []repository.Repository
	for _, repo := range repos {
		// If include list is specified, only include repos in that list
		if len(config.Include) > 0 {
			if includeMap[repo.Name] {
				filtered = append(filtered, repo)
			}
		} else {
			// If exclude list is specified, exclude repos in that list
			if !excludeMap[repo.Name] {
				filtered = append(filtered, repo)
			}
		}
	}

	return filtered
}