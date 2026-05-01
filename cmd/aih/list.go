package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/vaughngit/aih/internal/registry"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all agents in the central registry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := registry.List()
			if err != nil {
				return err
			}
			if len(ms) == 0 {
				dir, _ := registry.CentralDir()
				fmt.Fprintf(os.Stderr, "aih: no agents in %s\n", dir)
				return nil
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tKIND\tWORKSPACE\tDESCRIPTION")
			for _, m := range ms {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", m.Name, m.Kind, m.Workspace, m.Description)
			}
			return tw.Flush()
		},
	}
}
