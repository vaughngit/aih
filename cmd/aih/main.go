package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "aih: "+err.Error())
		exit(1)
	}
}

var exit = os.Exit

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "aih",
		Short:         "agentic-ai harness — launch any agent CLI from anywhere with the right context",
		Long:          "aih reads a TOML manifest and launches the configured agent backend (claude-code, codex, crush, kiro, opencode, generic) in the right workspace with the right env, hooks, and resources.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newLaunchCmd(),
		newListCmd(),
		newShowCmd(),
	)

	return root
}
