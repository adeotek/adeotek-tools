package filter

import (
	"reflect"
	"testing"

	"github.com/adeotek/git-multi-repo-clone/internal/config"
	"github.com/adeotek/git-multi-repo-clone/internal/repository"
)

func TestFilterRepositoriesWithNoFilters(t *testing.T) {
	// Setup test data
	repos := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo2", URL: "https://example.com/repo2"},
		{Name: "repo3", URL: "https://example.com/repo3"},
	}

	config := &config.Config{
		// No include or exclude lists
	}

	// Run the filter
	filtered := FilterRepositories(repos, config)

	// Verify no filtering happened
	if !reflect.DeepEqual(filtered, repos) {
		t.Errorf("Expected all repos to be returned when no filters are specified, got %v", filtered)
	}
}

func TestFilterRepositoriesWithIncludeList(t *testing.T) {
	// Setup test data
	repos := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo2", URL: "https://example.com/repo2"},
		{Name: "repo3", URL: "https://example.com/repo3"},
	}

	config := &config.Config{
		Include: []string{"repo1", "repo3"},
	}

	// Run the filter
	filtered := FilterRepositories(repos, config)

	// Verify only included repos are returned
	expected := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo3", URL: "https://example.com/repo3"},
	}

	if len(filtered) != len(expected) {
		t.Fatalf("Expected %d repos, got %d", len(expected), len(filtered))
	}

	// Check that the filtered list contains the expected repos
	for i, repo := range filtered {
		if repo.Name != expected[i].Name {
			t.Errorf("Expected repo at index %d to be %s, got %s", i, expected[i].Name, repo.Name)
		}
	}
}

func TestFilterRepositoriesWithExcludeList(t *testing.T) {
	// Setup test data
	repos := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo2", URL: "https://example.com/repo2"},
		{Name: "repo3", URL: "https://example.com/repo3"},
	}

	config := &config.Config{
		Exclude: []string{"repo2"},
	}

	// Run the filter
	filtered := FilterRepositories(repos, config)

	// Verify excluded repos are not returned
	expected := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo3", URL: "https://example.com/repo3"},
	}

	if len(filtered) != len(expected) {
		t.Fatalf("Expected %d repos, got %d", len(expected), len(filtered))
	}

	// Check that the filtered list contains the expected repos
	for i, repo := range filtered {
		if repo.Name != expected[i].Name {
			t.Errorf("Expected repo at index %d to be %s, got %s", i, expected[i].Name, repo.Name)
		}
	}
}

func TestFilterRepositoriesWithBothLists(t *testing.T) {
	// Setup test data
	repos := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo2", URL: "https://example.com/repo2"},
		{Name: "repo3", URL: "https://example.com/repo3"},
		{Name: "repo4", URL: "https://example.com/repo4"},
	}

	config := &config.Config{
		Include: []string{"repo1", "repo3"},
		Exclude: []string{"repo3", "repo4"}, // This should be ignored when include is specified
	}

	// Run the filter
	filtered := FilterRepositories(repos, config)

	// Verify only included repos are returned (exclude list should be ignored)
	expected := []repository.Repository{
		{Name: "repo1", URL: "https://example.com/repo1"},
		{Name: "repo3", URL: "https://example.com/repo3"},
	}

	if len(filtered) != len(expected) {
		t.Fatalf("Expected %d repos, got %d", len(expected), len(filtered))
	}

	// Check that the filtered list contains the expected repos
	for i, repo := range filtered {
		if repo.Name != expected[i].Name {
			t.Errorf("Expected repo at index %d to be %s, got %s", i, expected[i].Name, repo.Name)
		}
	}
}
