package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
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

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "aih: "+err.Error())
		os.Exit(1)
	}
}

func newLaunchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "launch <name>",
		Short: "Launch the named agent (Phase 1)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet — Phase 1")
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all agents in the central registry (Phase 1)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet — Phase 1")
		},
	}
}

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Print the resolved manifest for the named agent (Phase 1)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet — Phase 1")
		},
	}
}
