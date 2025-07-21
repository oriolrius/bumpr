package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/oriol/bumpr/internal/external"
	"github.com/oriol/bumpr/internal/release"
)

var (
	dryRun   bool
	source   string
	verbose  bool
	noPush   bool
	noCommit bool
	quiet    bool
	force    bool
)

var rootCmd = &cobra.Command{
	Use:   "bumpr",
	Short: "A cross-platform CLI tool for automated software releases",
	Long: `bumpr is a release automation tool that handles version bumping
and orchestrates git operations for creating releases.

It supports multiple version source files including pyproject.toml,
package.json, and .version files.`,
}

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Bump patch version (x.x.X)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRelease("patch")
	},
}

var minorCmd = &cobra.Command{
	Use:   "minor",
	Short: "Bump minor version (x.X.0)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRelease("minor")
	},
}

var majorCmd = &cobra.Command{
	Use:   "major",
	Short: "Bump major version (X.0.0)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRelease("major")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bumpr version %s (built at %s)\n", getVersion(), getBuildTime())
	},
}

var republishCmd = &cobra.Command{
	Use:   "republish",
	Short: "Republish the current version (removes and recreates tags/releases)",
	Long: `Republish the current version without incrementing it.
This command will:
- Remove existing local and remote tags
- Delete the GitHub release if it exists
- Recreate the tag and push it
- Create a new GitHub release`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRelease("republish")
	},
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without execution")
	flags.StringVarP(&source, "source", "s", "", "Specify version source file (auto-detect if not provided)")
	flags.BoolVarP(&verbose, "verbose", "v", false, "Show all executed commands")
	flags.BoolVar(&noPush, "no-push", false, "Skip pushing tags to remote repository")
	flags.BoolVar(&noCommit, "no-commit", false, "Skip committing changes")
	flags.BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	flags.BoolVarP(&force, "force", "f", false, "Skip safety checks and confirmations")

	rootCmd.AddCommand(patchCmd)
	rootCmd.AddCommand(minorCmd)
	rootCmd.AddCommand(majorCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(republishCmd)
}

func getVersion() string {
	return Version
}

func getBuildTime() string {
	return BuildTime
}

func Execute() error {
	return rootCmd.Execute()
}

func runRelease(bumpType string) error {
	if quiet && verbose {
		return fmt.Errorf("cannot use --quiet and --verbose together")
	}

	runner := external.NewRunner(verbose)
	orchestrator := release.NewOrchestrator(runner, verbose)

	options := release.Options{
		BumpType: bumpType,
		Source:   source,
		DryRun:   dryRun,
		Verbose:  verbose,
		NoPush:   noPush,
		NoCommit: noCommit,
		Quiet:    quiet,
		Force:    force,
	}

	return orchestrator.Execute(options)
}