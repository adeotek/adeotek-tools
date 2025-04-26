package filter

import (
	"reflect"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
)

func TestFilterRepositories(t *testing.T) {
	// Create test repositories
	testRepos := []repository.Repository{
		{
			Id:       1,
			Login:    "owner1",
			Name:     "repo1",
			FullName: "owner1/repo1",
			URL:      "https://example.com/owner1/repo1.git",
		},
		{
			Id:       2,
			Login:    "owner1",
			Name:     "repo2",
			FullName: "owner1/repo2",
			URL:      "https://example.com/owner1/repo2.git",
		},
		{
			Id:       3,
			Login:    "owner2",
			Name:     "repo3",
			FullName: "owner2/repo3",
			URL:      "https://example.com/owner2/repo3.git",
		},
	}

	// Test cases
	tests := []struct {
		name        string
		include     []string
		exclude     []string
		expectedIDs []int
	}{
		{
			name:        "No filters",
			include:     []string{},
			exclude:     []string{},
			expectedIDs: []int{1, 2, 3}, // Should return all repos
		},
		{
			name:        "Include filter",
			include:     []string{"owner1/repo1", "owner2/repo3"},
			exclude:     []string{},
			expectedIDs: []int{1, 3}, // Should return only owner1/repo1 and owner2/repo3
		},
		{
			name:        "Exclude filter",
			include:     []string{},
			exclude:     []string{"owner1/repo2"},
			expectedIDs: []int{1, 3}, // Should exclude owner1/repo2
		},
		{
			name:        "Include and exclude (include takes precedence)",
			include:     []string{"owner1/repo1"},
			exclude:     []string{"owner1/repo1", "owner1/repo2"},
			expectedIDs: []int{1}, // Include takes precedence, so only owner1/repo1 is included
		},
		{
			name:        "Include non-existent repo",
			include:     []string{"owner3/repo4"},
			exclude:     []string{},
			expectedIDs: []int{}, // No repos match the include filter, both slices are empty so they should match
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create provider config with include/exclude lists
			provider := &config.ProviderConfig{
				Include: tt.include,
				Exclude: tt.exclude,
			}

			// Filter repositories
			filtered := FilterRepositories(testRepos, provider, false)

			// Extract IDs for easier comparison
			var filteredIDs []int
			for _, repo := range filtered {
				filteredIDs = append(filteredIDs, repo.Id)
			}

			// Check if filtered repos match expected repos
			// Special case for empty slices to avoid test failure
			if len(filteredIDs) == 0 && len(tt.expectedIDs) == 0 {
				// Both are empty, so they match
				return
			}

			if !reflect.DeepEqual(filteredIDs, tt.expectedIDs) {
				t.Errorf("FilterRepositories() got = %v, want %v", filteredIDs, tt.expectedIDs)
			}
		})
	}

	// Test verbose mode (just to ensure it doesn't panic)
	provider := &config.ProviderConfig{
		Include: []string{"owner1/repo1"},
		Exclude: []string{"owner1/repo2"},
	}
	FilterRepositories(testRepos, provider, true)
}
