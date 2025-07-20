package external

import (
	"context"
)

type GitHubCommands struct {
	runner  CommandRunner
	verbose bool
}

func NewGitHubCommands(runner CommandRunner, verbose bool) *GitHubCommands {
	return &GitHubCommands{
		runner:  runner,
		verbose: verbose,
	}
}

func (g *GitHubCommands) IsAvailable() bool {
	_, err := g.runner.Run(context.Background(), "gh", "--version")
	return err == nil
}

func (g *GitHubCommands) CreateRelease(tagName, title, notes string) error {
	args := []string{"release", "create", tagName, "--title", title, "--notes", notes}
	_, err := g.runner.Run(context.Background(), "gh", args...)
	return err
}

func (g *GitHubCommands) CreateDraftRelease(tagName, title, notes string) error {
	args := []string{"release", "create", tagName, "--title", title, "--notes", notes, "--draft"}
	_, err := g.runner.Run(context.Background(), "gh", args...)
	return err
}

func (g *GitHubCommands) GetReleaseURL(tagName string) (string, error) {
	result, err := g.runner.RunWithOutput(context.Background(), "gh", "release", "view", tagName, "--json", "url", "-q", ".url")
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}