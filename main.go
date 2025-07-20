package main

import (
	"fmt"
	"os"

	"github.com/oriol/bumpr/cmd"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func init() {
	// Pass build-time variables to cmd package
	cmd.Version = Version
	cmd.BuildTime = BuildTime
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}