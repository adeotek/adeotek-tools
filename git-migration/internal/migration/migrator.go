package migration

import (
	"fmt"
	"path"
	"strings"

	"github.com/adeotek/adeotek-tools/git-migration/internal/config"
	"github.com/adeotek/adeotek-tools/git-migration/internal/forgejo"
	"github.com/adeotek/adeotek-tools/git-migration/internal/gitea"
)

// Result holds migration outcome counts.
type Result struct {
	Migrated int
	Skipped  int
	Failed   int
}

// Migrator orchestrates repo discovery, filtering, and migration.
type Migrator struct {
	gitea   *gitea.Client
	forgejo *forgejo.Client
	cfg     *config.Config
}

// New creates a Migrator.
func New(g *gitea.Client, f *forgejo.Client, cfg *config.Config) *Migrator {
	return &Migrator{gitea: g, forgejo: f, cfg: cfg}
}

// Run executes the full migration pipeline: enumerate → filter → (dry-run|migrate).
func (m *Migrator) Run() (Result, error) {
	repos, err := m.gitea.ListRepos(m.cfg.Orgs, m.cfg.Users)
	if err != nil {
		return Result{}, fmt.Errorf("enumerate repos: %w", err)
	}
	fmt.Printf("Found %d repo(s) in Gitea\n", len(repos))

	repos = m.filter(repos)
	fmt.Printf("%d repo(s) after filtering\n", len(repos))

	if m.cfg.DryRun {
		m.printDryRunPlan(repos)
		return Result{}, nil
	}

	var result Result
	giteaBase := strings.TrimRight(m.cfg.GiteaURL, "/")

	for _, repo := range repos {
		destOwner := m.destOwner(repo.Owner.Login)

		if err := m.forgejo.EnsureOrg(destOwner); err != nil {
			fmt.Printf("[ERROR] %s/%s: ensure org %q: %v\n", repo.Owner.Login, repo.Name, destOwner, err)
			result.Failed++
			continue
		}

		exists, err := m.forgejo.RepoExists(destOwner, repo.Name)
		if err != nil {
			fmt.Printf("[ERROR] %s/%s: check exists: %v\n", repo.Owner.Login, repo.Name, err)
			result.Failed++
			continue
		}

		if exists {
			switch m.cfg.OnConflict {
			case "skip":
				fmt.Printf("[SKIP]  %s/%s -> %s/%s (already exists)\n", repo.Owner.Login, repo.Name, destOwner, repo.Name)
				result.Skipped++
				continue
			case "fail":
				return result, fmt.Errorf("repo %s/%s already exists in Forgejo (use --on-conflict=remigrate to overwrite)", destOwner, repo.Name)
			case "remigrate":
				if err := m.forgejo.DeleteRepo(destOwner, repo.Name); err != nil {
					fmt.Printf("[ERROR] %s/%s: delete for remigrate: %v\n", repo.Owner.Login, repo.Name, err)
					result.Failed++
					continue
				}
				fmt.Printf("[REMIG] %s/%s -> %s/%s\n", repo.Owner.Login, repo.Name, destOwner, repo.Name)
			}
		}

		req := forgejo.MigrateRequest{
			AuthToken:    m.cfg.GiteaToken,
			CloneAddr:    fmt.Sprintf("%s/%s/%s", giteaBase, repo.Owner.Login, repo.Name),
			Description:  repo.Description,
			Issues:       m.cfg.MigrateIssues,
			Labels:       m.cfg.MigrateLabels,
			Milestones:   m.cfg.MigrateMilestones,
			Mirror:       false,
			Private:      repo.Private,
			PullRequests: m.cfg.MigratePullRequests,
			Releases:     m.cfg.MigrateReleases,
			RepoName:     repo.Name,
			RepoOwner:    destOwner,
			Wiki:         m.cfg.MigrateWiki,
			Service:      "gitea",
		}

		label := fmt.Sprintf("%s/%s -> %s/%s", repo.Owner.Login, repo.Name, destOwner, repo.Name)
		if err := m.migrateWithFallback(label, req); err != nil {
			fmt.Printf("[ERROR] %s: %v\n", label, err)
			result.Failed++
			continue
		}

		fmt.Printf("[OK]    %s/%s -> %s/%s\n", repo.Owner.Login, repo.Name, destOwner, repo.Name)
		result.Migrated++
	}

	return result, nil
}

func (m *Migrator) filter(repos []gitea.Repo) []gitea.Repo {
	excludeSet := make(map[string]bool, len(m.cfg.Exclude))
	for _, e := range m.cfg.Exclude {
		excludeSet[e] = true
	}

	var out []gitea.Repo
	for _, r := range repos {
		if excludeSet[r.Name] || excludeSet[r.FullName] {
			continue
		}
		if m.cfg.Filter != "" {
			matched, err := path.Match(m.cfg.Filter, r.Name)
			if err != nil || !matched {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// migrateWithFallback attempts migration and retries with individual features
// disabled whenever Forgejo reports that a feature is not available on the source.
// This handles repos where issue tracking, wiki, etc. are disabled in Gitea —
// Forgejo returns 500 "listing X: not found" instead of treating it as empty.
func (m *Migrator) migrateWithFallback(label string, req forgejo.MigrateRequest) error {
	for {
		err := m.forgejo.MigrateRepo(req)
		if err == nil {
			return nil
		}
		feature := disabledFeature(err.Error())
		if feature == "" {
			return err
		}
		fmt.Printf("[WARN]  %s: %s not available on source, skipping\n", label, feature)
		applyFeatureSkip(&req, feature)
	}
}

// disabledFeature inspects a Forgejo error message and returns the feature name
// that triggered a "not found" on the source, or "" if the error is something else.
func disabledFeature(errMsg string) string {
	if !strings.Contains(errMsg, "not found") {
		return ""
	}
	for _, f := range []struct{ pattern, name string }{
		{"listing issues", "issues"},
		{"listing labels", "labels"},
		{"listing milestone", "milestones"},
		{"listing pull", "pull requests"},
		{"listing release", "releases"},
		{"listing wiki", "wiki"},
	} {
		if strings.Contains(errMsg, f.pattern) {
			return f.name
		}
	}
	return ""
}

// applyFeatureSkip disables the named feature in req.
func applyFeatureSkip(req *forgejo.MigrateRequest, feature string) {
	switch feature {
	case "issues":
		req.Issues = false
	case "labels":
		req.Labels = false
	case "milestones":
		req.Milestones = false
	case "pull requests":
		req.PullRequests = false
	case "releases":
		req.Releases = false
	case "wiki":
		req.Wiki = false
	}
}

func (m *Migrator) destOwner(sourceOwner string) string {
	if dest, ok := m.cfg.OrgMappings[sourceOwner]; ok {
		return dest
	}
	return sourceOwner
}

func (m *Migrator) printDryRunPlan(repos []gitea.Repo) {
	fmt.Println("\n[DRY RUN] Migration plan:")
	fmt.Printf("  %-45s  %s\n", "Source", "Destination")
	fmt.Printf("  %-45s  %s\n", strings.Repeat("-", 45), strings.Repeat("-", 30))
	for _, r := range repos {
		dest := m.destOwner(r.Owner.Login)
		fmt.Printf("  %-45s  %s/%s\n", r.FullName, dest, r.Name)
	}
	fmt.Printf("\n  Total: %d repo(s) would be migrated\n", len(repos))
}
