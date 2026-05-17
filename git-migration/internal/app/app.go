package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/adeotek/adeotek-tools/git-migration/internal/config"
	"github.com/adeotek/adeotek-tools/git-migration/internal/forgejo"
	"github.com/adeotek/adeotek-tools/git-migration/internal/gitea"
	"github.com/adeotek/adeotek-tools/git-migration/internal/migration"
)

const Version = "0.1.0"

// Run is the application entry point.
func Run() {
	giteaURL := flag.String("gitea-url", "", "Gitea server URL (required)")
	giteaToken := flag.String("gitea-token", "", "Gitea API token (or set GITEA_TOKEN)")
	forgejoURL := flag.String("forgejo-url", "", "Forgejo server URL (required)")
	forgejoToken := flag.String("forgejo-token", "", "Forgejo API token (or set FORGEJO_TOKEN)")
	orgs := flag.String("orgs", "", "Comma-separated Gitea orgs to migrate")
	users := flag.String("users", "", "Comma-separated Gitea users to migrate")
	filter := flag.String("filter", "", "Glob pattern to filter repo names (e.g. infra-*)")
	exclude := flag.String("exclude", "", "Comma-separated repo names or full names to exclude")
	mappings := config.NewOrgMappings()
	flag.Var(mappings, "map-org", "Map source org to destination org (format: src:dst, repeatable)")
	onConflict := flag.String("on-conflict", "skip", "Behaviour when repo already exists: skip|fail|remigrate")
	dryRun := flag.Bool("dry-run", false, "Print migration plan without making changes")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	showVersion := flag.Bool("version", false, "Print version and exit")
	showHelp := flag.Bool("help", false, "Print usage and exit")
	skipIssues := flag.Bool("skip-issues", false, "Skip migrating issues")
	skipLabels := flag.Bool("skip-labels", false, "Skip migrating labels")
	skipMilestones := flag.Bool("skip-milestones", false, "Skip migrating milestones")
	skipPullRequests := flag.Bool("skip-pull-requests", false, "Skip migrating pull requests")
	skipReleases := flag.Bool("skip-releases", false, "Skip migrating releases")
	skipWiki := flag.Bool("skip-wiki", false, "Skip migrating wiki")

	flag.Parse()

	fmt.Printf("git-migration version %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

	if *showVersion {
		return
	}
	if *showHelp {
		PrintUsage()
		return
	}

	cfg := &config.Config{
		GiteaURL:            *giteaURL,
		GiteaToken:          config.ResolveToken(*giteaToken, "GITEA_TOKEN"),
		ForgejoURL:          *forgejoURL,
		ForgejoToken:        config.ResolveToken(*forgejoToken, "FORGEJO_TOKEN"),
		Orgs:                config.SplitList(*orgs),
		Users:               config.SplitList(*users),
		Filter:              *filter,
		Exclude:             config.SplitList(*exclude),
		OrgMappings:         mappings,
		OnConflict:          *onConflict,
		DryRun:              *dryRun,
		Verbose:             *verbose,
		MigrateIssues:       !*skipIssues,
		MigrateLabels:       !*skipLabels,
		MigrateMilestones:   !*skipMilestones,
		MigratePullRequests: !*skipPullRequests,
		MigrateReleases:     !*skipReleases,
		MigrateWiki:         !*skipWiki,
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// verbose prints configuration details before the migration run begins.
	// Migration output ([OK], [SKIP], etc.) is always shown.
	if cfg.Verbose {
		fmt.Printf("Gitea:      %s\n", cfg.GiteaURL)
		fmt.Printf("Forgejo:    %s\n", cfg.ForgejoURL)
		if len(cfg.Orgs) > 0 {
			fmt.Printf("Orgs:       %v\n", cfg.Orgs)
		}
		if len(cfg.Users) > 0 {
			fmt.Printf("Users:      %v\n", cfg.Users)
		}
		if cfg.Filter != "" {
			fmt.Printf("Filter:     %s\n", cfg.Filter)
		}
		if len(cfg.Exclude) > 0 {
			fmt.Printf("Exclude:    %v\n", cfg.Exclude)
		}
		if len(cfg.OrgMappings) > 0 {
			fmt.Printf("Org maps:   %s\n", cfg.OrgMappings)
		}
		fmt.Printf("OnConflict: %s\n", cfg.OnConflict)
	}

	giteaClient := gitea.NewClient(cfg.GiteaURL, cfg.GiteaToken)
	forgejoClient := forgejo.NewClient(cfg.ForgejoURL, cfg.ForgejoToken)
	migrator := migration.New(giteaClient, forgejoClient, cfg)

	result, err := migrator.Run()
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	if !cfg.DryRun {
		fmt.Printf("\nDone: %d migrated, %d skipped, %d failed\n", result.Migrated, result.Skipped, result.Failed)
		if result.Failed > 0 {
			os.Exit(1)
		}
	}
}

// printFlagDefaults prints all registered flags with -- prefix.
// Replaces flag.PrintDefaults() which always renders single-dash output.
func printFlagDefaults() {
	flag.VisitAll(func(f *flag.Flag) {
		isBool := f.DefValue == "true" || f.DefValue == "false"
		if isBool {
			fmt.Printf("  --%s\n", f.Name)
		} else {
			fmt.Printf("  --%s string\n", f.Name)
		}
		if f.DefValue != "" && !isBool {
			fmt.Printf("\t%s (default %q)\n", f.Usage, f.DefValue)
		} else {
			fmt.Printf("\t%s\n", f.Usage)
		}
	})
}

// PrintUsage prints CLI usage.
func PrintUsage() {
	fmt.Println("git-migration - Migrate repositories from Gitea to Forgejo")
	fmt.Println("\nUsage:")
	fmt.Println("  git-migration [flags]")
	fmt.Println("\nFlags:")
	printFlagDefaults()
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  GITEA_TOKEN    Gitea API token (overridden by --gitea-token)")
	fmt.Println("  FORGEJO_TOKEN  Forgejo API token (overridden by --forgejo-token)")
	fmt.Println("\nExamples:")
	fmt.Println("  # Dry run: preview what would be migrated from org 'myorg'")
	fmt.Println("  git-migration --gitea-url https://gitea.lan --forgejo-url https://forgejo.lan \\")
	fmt.Println("    --orgs myorg --dry-run")
	fmt.Println()
	fmt.Println("  # Migrate with org mapping, skip existing repos")
	fmt.Println("  git-migration --gitea-url https://gitea.lan --forgejo-url https://forgejo.lan \\")
	fmt.Println("    --orgs src-org --map-org src-org:dest-org --on-conflict skip")
	fmt.Println()
	fmt.Println("  # Migrate only repos matching a pattern, excluding one")
	fmt.Println("  git-migration --gitea-url https://gitea.lan --forgejo-url https://forgejo.lan \\")
	fmt.Println("    --orgs myorg --filter 'infra-*' --exclude infra-legacy")
}
