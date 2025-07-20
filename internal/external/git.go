package external

import (
	"context"
	"strings"
)

type GitCommands struct {
	runner  CommandRunner
	verbose bool
}

func NewGitCommands(runner CommandRunner, verbose bool) *GitCommands {
	return &GitCommands{
		runner:  runner,
		verbose: verbose,
	}
}

func (g *GitCommands) Add(files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := g.runner.Run(context.Background(), "git", args...)
	return err
}

func (g *GitCommands) Commit(message string) error {
	args := []string{"commit", "-m", message}
	_, err := g.runner.Run(context.Background(), "git", args...)
	return err
}

func (g *GitCommands) CreateTag(tagName, message string) error {
	args := []string{"tag", "-a", tagName, "-m", message}
	_, err := g.runner.Run(context.Background(), "git", args...)
	return err
}

func (g *GitCommands) PushTag(tagName string) error {
	args := []string{"push", "origin", tagName}
	_, err := g.runner.Run(context.Background(), "git", args...)
	return err
}

func (g *GitCommands) PushTagWithForce(tagName string) error {
	args := []string{"push", "origin", tagName, "--force"}
	_, err := g.runner.Run(context.Background(), "git", args...)
	return err
}

func (g *GitCommands) Push(branch string) error {
	args := []string{"push", "origin", branch}
	_, err := g.runner.Run(context.Background(), "git", args...)
	return err
}

func (g *GitCommands) DeleteLocalTag(tagName string) error {
	args := []string{"tag", "-d", tagName}
	// Ignore errors for this operation
	g.runner.Run(context.Background(), "git", args...)
	return nil
}

func (g *GitCommands) DeleteRemoteTag(tagName string) error {
	args := []string{"push", "origin", "--delete", tagName}
	// Ignore errors for this operation
	g.runner.Run(context.Background(), "git", args...)
	return nil
}

func (g *GitCommands) Status() (string, error) {
	result, err := g.runner.RunWithOutput(context.Background(), "git", "status", "--porcelain")
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func (g *GitCommands) IsClean() (bool, error) {
	status, err := g.Status()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(status) == "", nil
}

func (g *GitCommands) CurrentBranch() (string, error) {
	result, err := g.runner.RunWithOutput(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

func (g *GitCommands) IsRepository() bool {
	_, err := g.runner.Run(context.Background(), "git", "rev-parse", "--git-dir")
	return err == nil
}

func (g *GitCommands) TagExists(tagName string) bool {
	_, err := g.runner.Run(context.Background(), "git", "rev-parse", tagName)
	return err == nil
}