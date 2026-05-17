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
				fmt.Printf("[REMIG] %s/%s -> %s/%s\n", repo.Owner.Login, repo.Name, destOwner, repo.Name)
			}
		}

		req := forgejo.MigrateRequest{
			AuthToken:    m.cfg.GiteaToken,
			CloneAddr:    fmt.Sprintf("%s/%s/%s", giteaBase, repo.Owner.Login, repo.Name),
			Description:  repo.Description,
			Issues:       true,
			Labels:       true,
			Milestones:   true,
			Mirror:       false,
			Private:      repo.Private,
			PullRequests: true,
			Releases:     true,
			RepoName:     repo.Name,
			RepoOwner:    destOwner,
			Wiki:         true,
			Service:      "gitea",
		}

		if err := m.forgejo.MigrateRepo(req); err != nil {
			fmt.Printf("[ERROR] %s/%s -> %s/%s: %v\n", repo.Owner.Login, repo.Name, destOwner, repo.Name, err)
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
