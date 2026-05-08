package main

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"github.com/vaughngit/aih/internal/registry"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Print the resolved manifest for the named agent (paths expanded)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := registry.LoadCentral(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "# source: %s\n", m.Source)
			enc := toml.NewEncoder(cmd.OutOrStdout())
			enc.SetIndentTables(true)
			return enc.Encode(m)
		},
	}
}
