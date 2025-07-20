package external

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CommandRunner interface {
	Run(ctx context.Context, cmd string, args ...string) (*CommandResult, error)
	RunWithOutput(ctx context.Context, cmd string, args ...string) (*CommandResult, error)
}

type CommandResult struct {
	Command  string
	Args     []string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

type DefaultRunner struct {
	verbose bool
}

func NewRunner(verbose bool) CommandRunner {
	return &DefaultRunner{verbose: verbose}
}

func (r *DefaultRunner) Run(ctx context.Context, cmd string, args ...string) (*CommandResult, error) {
	return r.execute(ctx, false, cmd, args...)
}

func (r *DefaultRunner) RunWithOutput(ctx context.Context, cmd string, args ...string) (*CommandResult, error) {
	return r.execute(ctx, true, cmd, args...)
}

func (r *DefaultRunner) execute(ctx context.Context, captureOutput bool, cmd string, args ...string) (*CommandResult, error) {
	if r.verbose {
		fmt.Printf("→ %s %s\n", cmd, strings.Join(args, " "))
	}

	start := time.Now()
	command := exec.CommandContext(ctx, cmd, args...)
	
	var stdout, stderr bytes.Buffer
	if captureOutput {
		command.Stdout = &stdout
		command.Stderr = &stderr
	} else if r.verbose {
		command.Stdout = &prefixedWriter{prefix: "  "}
		command.Stderr = &prefixedWriter{prefix: "  "}
	}

	err := command.Run()
	duration := time.Since(start)

	result := &CommandResult{
		Command:  cmd,
		Args:     args,
		ExitCode: 0,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
		return result, fmt.Errorf("command failed: %s %s: %w", cmd, strings.Join(args, " "), err)
	}

	return result, nil
}

type prefixedWriter struct {
	prefix string
}

func (w *prefixedWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		if line != "" || i < len(lines)-1 {
			fmt.Printf("%s%s", w.prefix, line)
			if i < len(lines)-1 {
				fmt.Println()
			}
		}
	}
	return len(p), nil
}