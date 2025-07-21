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

	// Calculate new version
	var newVersion string
	if options.BumpType == "republish" {
		// For republish, use the current version
		newVersion = currentVersion
	} else {
		// Parse bump type
		bumpType, err := version.ParseBumpType(options.BumpType)
		if err != nil {
			return err
		}
		
		newVersion, err = version.Bump(currentVersion, bumpType)
		if err != nil {
			return fmt.Errorf("failed to bump version: %w", err)
		}
	}

	if !options.Quiet {
		if options.BumpType == "republish" {
			fmt.Printf("🔄 Republishing version: %s\n\n", currentVersion)
		} else {
			fmt.Printf("📊 Current version: %s\n", currentVersion)
			fmt.Printf("📈 New version: %s\n\n", newVersion)
		}
	}

	if options.DryRun {
		o.showDryRunCommands(sourceFile, newVersion, options)
		return nil
	}

	// Update version file (skip for republish)
	if options.BumpType != "republish" {
		if err := source.SetVersion(sourceFile, newVersion); err != nil {
			return fmt.Errorf("failed to update version: %w", err)
		}

		if !options.Quiet {
			fmt.Printf("✅ Updated %s with new version\n", filepath.Base(sourceFile))
		}
	}

	// Git operations (skip commit for republish)
	if !options.NoCommit && options.BumpType != "republish" {
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

		// Push commit to origin
		if !options.NoPush {
			branch, err := o.gitCmd.CurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			if err := o.gitCmd.Push(branch); err != nil {
				return fmt.Errorf("failed to push commit: %w", err)
			}

			if !options.Quiet {
				fmt.Printf("📤 Pushed commit to origin/%s\n", branch)
			}
		}
	}

	// Tag operations
	if options.BumpType == "republish" {
		// For republish, we need to be more aggressive with cleanup
		if err := o.forceCleanupTagAndRelease(newVersion, options); err != nil {
			if !options.Quiet {
				fmt.Printf("⚠️  Warning: failed to cleanup existing tag/release: %v\n", err)
			}
		}
	} else {
		if err := o.cleanupExistingTag(newVersion, options); err != nil {
			if !options.Quiet {
				fmt.Printf("⚠️  Warning: failed to cleanup existing tag: %v\n", err)
			}
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
		// Force push the tag to ensure it's updated if it already existed
		if err := o.gitCmd.PushTagWithForce(newVersion); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}

		if !options.Quiet {
			fmt.Printf("📤 Pushed tag: %s (forced)\n", newVersion)
		}

		// Create GitHub release
		if o.githubCmd.IsAvailable() {
			if err := o.createGitHubRelease(newVersion, options); err != nil {
				if !options.Quiet {
					fmt.Printf("⚠️  Warning: failed to create GitHub release: %v\n", err)
					fmt.Println("   The tag has been pushed, so the workflow will still run.")
				}
			} else if !options.Quiet {
				fmt.Printf("🎉 Created GitHub release for %s\n", newVersion)
			}
		} else if !options.Quiet {
			fmt.Println("ℹ️  GitHub CLI (gh) not found. Skipping release creation.")
			fmt.Println("   Install it with: https://cli.github.com/")
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
	
	if options.BumpType == "republish" {
		if !options.NoPush && o.githubCmd.IsAvailable() {
			fmt.Printf("→ gh release delete %s --yes (if exists)\n", newVersion)
		}
	} else {
		fmt.Printf("→ Update %s with version %s\n", filepath.Base(sourceFile), newVersion)
		
		if !options.NoCommit {
			fmt.Printf("→ git add %s\n", sourceFile)
			fmt.Printf("→ git commit -m \"releasing %s\"\n", newVersion)
			
			if !options.NoPush {
				fmt.Println("→ git push origin <current-branch>")
			}
		}
	}
	
	// Tag cleanup if exists
	fmt.Printf("→ git tag -d %s (if exists)\n", newVersion)
	if !options.NoPush {
		fmt.Printf("→ git push origin --delete %s (if exists)\n", newVersion)
	}
	
	fmt.Printf("→ git tag -a %s -m \"Release: %s\"\n", newVersion, newVersion)
	
	if !options.NoPush {
		fmt.Printf("→ git push origin %s --force\n", newVersion)
		
		if o.githubCmd.IsAvailable() {
			fmt.Printf("→ gh release create %s --title \"Release %s\" --notes \"...\"\n", newVersion, newVersion)
		}
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

	if options.NoPush {
		fmt.Println("Next steps:")
		
		if !options.NoCommit {
			branch, _ := o.gitCmd.CurrentBranch()
			fmt.Printf("1. Push the commit when ready: git push origin %s\n", branch)
		}
		
		fmt.Printf("2. Push the tag when ready: git push origin %s --force\n", newVersion)
		fmt.Println("3. Create a GitHub release manually or run: gh release create " + newVersion)
	} else {
		// Everything was pushed automatically
		fmt.Println("The release process is complete!")
		fmt.Println()
		fmt.Println("GitHub Actions is now building your release. You can:")
		fmt.Println("- Check the build progress in GitHub Actions")
		fmt.Printf("- View the release at: https://github.com/USERNAME/REPO/releases/tag/%s\n", newVersion)
	}
}

func (o *Orchestrator) createGitHubRelease(version string, options Options) error {
	title := fmt.Sprintf("Release %s", version)
	notes := fmt.Sprintf("## Release %s\n\nAutomated release created by bumpr.", version)
	
	return o.githubCmd.CreateRelease(version, title, notes)
}

func (o *Orchestrator) forceCleanupTagAndRelease(tagName string, options Options) error {
	if options.Verbose && !options.Quiet {
		fmt.Printf("🧹 Force cleaning up tag and release %s...\n", tagName)
	}
	
	// Delete GitHub release first if gh is available
	if !options.NoPush && o.githubCmd.IsAvailable() {
		if err := o.githubCmd.DeleteRelease(tagName); err != nil {
			if options.Verbose && !options.Quiet {
				fmt.Printf("   Failed to delete GitHub release: %v\n", err)
			}
		} else if !options.Quiet {
			fmt.Printf("🗑️  Deleted GitHub release %s\n", tagName)
		}
	}
	
	// Delete local tag
	o.gitCmd.DeleteLocalTag(tagName)
	
	// Delete remote tag
	if !options.NoPush {
		o.gitCmd.DeleteRemoteTag(tagName)
	}
	
	return nil
}