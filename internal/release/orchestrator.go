package release

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/oriol/bumpr/internal/external"
	"github.com/oriol/bumpr/internal/sources"
	"github.com/oriol/bumpr/internal/version"
)

type Options struct {
	BumpType string
	Source   string
	DryRun   bool
	Verbose  bool
	NoPush   bool
	NoCommit bool
	Quiet    bool
	Force    bool
}

type Orchestrator struct {
	detector   *sources.Detector
	gitCmd     *external.GitCommands
	githubCmd  *external.GitHubCommands
	checker    *external.DependencyChecker
}

func NewOrchestrator(runner external.CommandRunner, verbose bool) *Orchestrator {
	return &Orchestrator{
		detector:  sources.NewDetector(),
		gitCmd:    external.NewGitCommands(runner, verbose),
		githubCmd: external.NewGitHubCommands(runner, verbose),
		checker:   external.NewDependencyChecker(runner),
	}
}

func (o *Orchestrator) Execute(options Options) error {
	if !options.Quiet {
		fmt.Printf("🚀 Starting release process...\n\n")
	}

	// Pre-flight checks
	if !options.Force {
		if err := o.runPreflightChecks(options); err != nil {
			return fmt.Errorf("pre-flight check failed: %w", err)
		}
	}

	// Detect or use specified version source
	source, sourceFile, err := o.detectVersionSource(options.Source)
	if err != nil {
		return err
	}

	if !options.Quiet {
		fmt.Printf("📄 Using version source: %s\n", sourceFile)
	}

	// Get current version
	currentVersion, err := source.GetVersion(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Parse bump type
	bumpType, err := version.ParseBumpType(options.BumpType)
	if err != nil {
		return err
	}

	// Calculate new version
	newVersion, err := version.Bump(currentVersion, bumpType)
	if err != nil {
		return fmt.Errorf("failed to bump version: %w", err)
	}

	if !options.Quiet {
		fmt.Printf("📊 Current version: %s\n", currentVersion)
		fmt.Printf("📈 New version: %s\n\n", newVersion)
	}

	if options.DryRun {
		o.showDryRunCommands(sourceFile, newVersion, options)
		return nil
	}

	// Update version file
	if err := source.SetVersion(sourceFile, newVersion); err != nil {
		return fmt.Errorf("failed to update version: %w", err)
	}

	if !options.Quiet {
		fmt.Printf("✅ Updated %s with new version\n", filepath.Base(sourceFile))
	}

	// Git operations
	if !options.NoCommit {
		if err := o.gitCmd.Add(sourceFile); err != nil {
			return fmt.Errorf("failed to stage file: %w", err)
		}

		commitMessage := fmt.Sprintf("releasing %s", newVersion)
		if err := o.gitCmd.Commit(commitMessage); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		if !options.Quiet {
			fmt.Printf("💾 Committed: %s\n", commitMessage)
		}
	}

	// Tag operations
	if err := o.cleanupExistingTag(newVersion, options); err != nil {
		if !options.Quiet {
			fmt.Printf("⚠️  Warning: failed to cleanup existing tag: %v\n", err)
		}
	}

	tagMessage := fmt.Sprintf("Release: %s", newVersion)
	if err := o.gitCmd.CreateTag(newVersion, tagMessage); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	if !options.Quiet {
		fmt.Printf("🏷️  Created tag: %s\n", newVersion)
	}

	if !options.NoPush {
		if err := o.gitCmd.PushTag(newVersion); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}

		if !options.Quiet {
			fmt.Printf("📤 Pushed tag: %s\n", newVersion)
		}
	}

	// Success message and next steps
	o.showSuccessMessage(newVersion, options)

	return nil
}

func (o *Orchestrator) runPreflightChecks(options Options) error {
	if options.Verbose && !options.Quiet {
		fmt.Println("🔍 Running pre-flight checks...")
	}

	skipCleanCheck := options.Force || options.NoCommit
	if err := o.checker.RunPreflightChecks(skipCleanCheck); err != nil {
		return err
	}

	if options.Verbose && !options.Quiet {
		fmt.Println("✅ Pre-flight checks passed")
		fmt.Println()
	}

	return nil
}

func (o *Orchestrator) detectVersionSource(sourceFile string) (sources.VersionSource, string, error) {
	if sourceFile != "" {
		// User specified a source file
		absPath, err := filepath.Abs(sourceFile)
		if err != nil {
			return nil, "", fmt.Errorf("invalid source file path: %w", err)
		}

		source, err := o.detector.GetSourceByFile(absPath)
		if err != nil {
			return nil, "", err
		}

		return source, absPath, nil
	}

	// Auto-detect
	cwd := "."
	source, filePath, err := o.detector.DetectSource(cwd)
	if err != nil {
		availableSources := o.detector.ListAvailableSources()
		return nil, "", fmt.Errorf("no version source file found. Supported files: %s", strings.Join(availableSources, ", "))
	}

	return source, filePath, nil
}

func (o *Orchestrator) cleanupExistingTag(tagName string, options Options) error {
	if o.gitCmd.TagExists(tagName) {
		if options.Verbose && !options.Quiet {
			fmt.Printf("🧹 Cleaning up existing tag %s...\n", tagName)
		}
		o.gitCmd.DeleteLocalTag(tagName)
		if !options.NoPush {
			o.gitCmd.DeleteRemoteTag(tagName)
		}
	}
	return nil
}

func (o *Orchestrator) showDryRunCommands(sourceFile, newVersion string, options Options) {
	fmt.Println("🔍 Dry run mode - commands that would be executed:")
	fmt.Println()
	
	fmt.Printf("→ Update %s with version %s\n", filepath.Base(sourceFile), newVersion)
	
	if !options.NoCommit {
		fmt.Printf("→ git add %s\n", sourceFile)
		fmt.Printf("→ git commit -m \"releasing %s\"\n", newVersion)
	}
	
	fmt.Printf("→ git tag -a %s -m \"Release: %s\"\n", newVersion, newVersion)
	
	if !options.NoPush {
		fmt.Printf("→ git push origin %s\n", newVersion)
	}
	
	fmt.Println()
	fmt.Println("Run without --dry-run to execute these commands.")
}

func (o *Orchestrator) showSuccessMessage(newVersion string, options Options) {
	if options.Quiet {
		return
	}

	fmt.Println()
	fmt.Printf("✅ Successfully released %s\n", newVersion)
	fmt.Println()
	fmt.Println("Next steps:")
	
	if !options.NoCommit && options.NoPush {
		fmt.Println("1. Push the commit to origin: git push origin")
	} else if !options.NoCommit {
		branch, _ := o.gitCmd.CurrentBranch()
		fmt.Printf("1. The commit has been created. Push when ready: git push origin %s\n", branch)
	}
	
	if options.NoPush {
		fmt.Printf("2. Push the tag when ready: git push origin %s\n", newVersion)
	}
	
	fmt.Println("3. Check GitHub Actions for the automated build")
	fmt.Printf("4. Once build completes, the release will be available at:\n")
	fmt.Printf("   https://github.com/USERNAME/REPO/releases/tag/%s\n", newVersion)
}