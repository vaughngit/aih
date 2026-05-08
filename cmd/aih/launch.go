package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vaughngit/aih/internal/backends"
	"github.com/vaughngit/aih/internal/registry"
	"github.com/vaughngit/aih/internal/runtime"
)

func newLaunchCmd() *cobra.Command {
	var (
		resume      bool
		addDirs     []string
		backendOver string
	)
	cmd := &cobra.Command{
		Use:   "launch <name> [-- <passthrough args...>]",
		Short: "Launch the named agent",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			passthru := args[1:]

			m, err := registry.LoadCentral(name)
			if err != nil {
				return err
			}
			if err := m.Validate(); err != nil {
				return err
			}

			kind := m.Kind
			if backendOver != "" {
				kind = backendOver
			}
			if kind == "" {
				return fmt.Errorf("manifest %q has no `kind` and no --backend override", name)
			}

			b, err := backends.Get(kind)
			if err != nil {
				return err
			}
			if err := b.ValidateConfig(m.BackendConfig(b.Name())); err != nil {
				return err
			}

			if resume && !b.SupportsResume() {
				return fmt.Errorf("backend %q does not support --resume", b.Name())
			}
			if len(addDirs) > 0 && !b.SupportsAddDir() {
				return fmt.Errorf("backend %q does not support --add", b.Name())
			}

			argv, err := b.BuildCommand(m, backends.LaunchOpts{
				Resume:    resume,
				AddDirs:   addDirs,
				PassThru:  passthru,
				BackendID: kind,
			})
			if err != nil {
				return err
			}

			res, err := runtime.Launch(m, argv)
			if err != nil {
				return err
			}
			if res.Signal != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "aih: child terminated by signal %s\n", res.Signal)
			}
			if res.ExitCode != 0 {
				exit(res.ExitCode)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&resume, "resume", false, "ask the backend to resume its previous session (where supported)")
	cmd.Flags().StringSliceVar(&addDirs, "add", nil, "extra resource path for this run only (where the backend supports it)")
	cmd.Flags().StringVar(&backendOver, "backend", "", "override the manifest's `kind` for this run (e.g. --backend generic)")
	return cmd
}
