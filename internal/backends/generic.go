package backends

import (
	"fmt"

	"github.com/vaughngit/aih/internal/manifest"
)

func init() { Register(&Generic{}) }

// Generic launches whatever Command the manifest declares verbatim. No flag
// magic. Pass-through args appended.
type Generic struct{}

func (*Generic) Name() string             { return "generic" }
func (*Generic) SupportsResume() bool     { return false }
func (*Generic) SupportsAddDir() bool     { return false }
func (*Generic) DefaultLogSubdir() string { return "generic" }

func (*Generic) BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error) {
	if len(m.Command) == 0 {
		return nil, fmt.Errorf("generic backend: manifest %q has no command (must be a non-empty array)", m.Name)
	}
	if opts.Resume {
		return nil, fmt.Errorf("generic backend: --resume not supported (manifest %q)", m.Name)
	}
	if len(opts.AddDirs) > 0 {
		return nil, fmt.Errorf("generic backend: --add not supported (manifest %q)", m.Name)
	}
	out := append([]string(nil), m.Command...)
	out = append(out, opts.PassThru...)
	return out, nil
}

func (*Generic) ValidateConfig(raw map[string]any) error {
	if len(raw) > 0 {
		return fmt.Errorf("generic backend: [backend.generic] block is not allowed (got keys %v)", keys(raw))
	}
	return nil
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
