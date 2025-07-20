package external

import (
	"context"
	"fmt"
	"strings"
)

type DependencyChecker struct {
	runner CommandRunner
}

func NewDependencyChecker(runner CommandRunner) *DependencyChecker {
	return &DependencyChecker{
		runner: runner,
	}
}

func (d *DependencyChecker) CheckGit() error {
	_, err := d.runner.Run(context.Background(), "git", "--version")
	if err != nil {
		return fmt.Errorf("git is not available in PATH. Please install git")
	}
	return nil
}

func (d *DependencyChecker) CheckGitHub() error {
	_, err := d.runner.Run(context.Background(), "gh", "--version")
	if err != nil {
		return fmt.Errorf("gh (GitHub CLI) is not available in PATH. Install from: https://cli.github.com/")
	}
	return nil
}

func (d *DependencyChecker) CheckRepository() error {
	_, err := d.runner.Run(context.Background(), "git", "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("not in a git repository")
	}
	return nil
}

func (d *DependencyChecker) CheckWorkingDirectory() error {
	result, err := d.runner.RunWithOutput(context.Background(), "git", "status", "--porcelain")
	if err != nil {
		return err
	}
	
	if strings.TrimSpace(result.Stdout) != "" {
		return fmt.Errorf("working directory is not clean. Please commit or stash changes")
	}
	
	return nil
}

func (d *DependencyChecker) RunPreflightChecks(skipCleanCheck bool) error {
	// Check git is available
	if err := d.CheckGit(); err != nil {
		return err
	}

	// Check we're in a git repository
	if err := d.CheckRepository(); err != nil {
		return err
	}

	// Check working directory is clean
	if !skipCleanCheck {
		if err := d.CheckWorkingDirectory(); err != nil {
			return err
		}
	}

	// GitHub CLI is optional, so we don't fail if it's not available
	// We'll just note it for later
	
	return nil
}