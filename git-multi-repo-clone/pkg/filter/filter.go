// Package filter provides repository filtering functionality
package filter

import (
	"fmt"
	"strings"

	"github.com/adeotek/git-multi-repo-clone/internal/config"
	"github.com/adeotek/git-multi-repo-clone/internal/repository"
)

// FilterRepositories applies include and exclude filters from config
func FilterRepositories(repos []repository.Repository, config *config.Config, verbose bool) []repository.Repository {
	// If neither include nor exclude is specified, return all repos
	if len(config.Include) == 0 && len(config.Exclude) == 0 {
		return repos
	}

	// Create a map for faster lookup
	includeMap := make(map[string]bool)
	for _, name := range config.Include {
		includeMap[name] = true
	}

	if verbose {
		includeList := make([]string, 0, len(includeMap))
		for name := range includeMap {
			includeList = append(includeList, name)
		}
		if len(includeMap) == 0 {
			fmt.Printf("----> `include` filter is empty\n")
		} else {
			fmt.Printf("----> `include` filter: %s\n", strings.Join(includeList, ", "))
		}
	}

	excludeMap := make(map[string]bool)
	for _, name := range config.Exclude {
		excludeMap[name] = true
	}

	if verbose {
		excludeList := make([]string, 0, len(excludeMap))
		for name := range excludeMap {
			excludeList = append(excludeList, name)
		}
		if len(excludeList) == 0 {
			fmt.Printf("----> `exclude` filter is empty\n")
		} else {
			fmt.Printf("----> `exclude` filter: %s\n", strings.Join(excludeList, ", "))
		}
	}

	var filtered []repository.Repository
	for _, repo := range repos {
		// if verbose {
		// 	fmt.Printf("----> applying filters for repo: %s\n", repo.FullName)
		// }

		// If include list is specified, only include repos in that list
		if len(config.Include) > 0 {
			if includeMap[repo.FullName] {
				filtered = append(filtered, repo)
			}
		} else {
			// If exclude list is specified, exclude repos in that list
			if !excludeMap[repo.FullName] {
				filtered = append(filtered, repo)
			}
		}
	}

	return filtered
}
