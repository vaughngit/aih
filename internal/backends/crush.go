package backends

import (
	"github.com/vaughngit/aih/internal/manifest"
)

func init() { Register(&Crush{}) }

// Crush wraps the Charm `crush` CLI.
//
// Resume semantics: aih's --resume becomes `crush --continue` (most recent
// session). For a specific session id use passthru: `aih launch x -- -s <id>`.
//
// No add-dir support: crush has no flag for additional context paths.
// `manifest.Resources` is silently ignored.
type Crush struct{}

func (*Crush) Name() string             { return "crush" }
func (*Crush) SupportsResume() bool     { return true }
func (*Crush) SupportsAddDir() bool     { return false }
func (*Crush) DefaultLogSubdir() string { return "crush" }

var crushAllowedKeys = map[string]keyType{
	"binary":   keyString,
	"data_dir": keyString,
	"yolo":     keyBool,
	"debug":    keyBool,
	"host":     keyString,
}

func (c *Crush) BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error) {
	cfg := m.BackendConfig(c.Name())
	argv := []string{stringOrDefault(cfg, "binary", "crush")}

	if opts.Resume {
		argv = append(argv, "--continue")
	}
	if v := stringOrDefault(cfg, "data_dir", ""); v != "" {
		argv = append(argv, "--data-dir", v)
	}
	if boolOrDefault(cfg, "yolo", false) {
		argv = append(argv, "--yolo")
	}
	if boolOrDefault(cfg, "debug", false) {
		argv = append(argv, "--debug")
	}
	if v := stringOrDefault(cfg, "host", ""); v != "" {
		argv = append(argv, "--host", v)
	}

	argv = append(argv, opts.PassThru...)
	return argv, nil
}

func (c *Crush) ValidateConfig(raw map[string]any) error {
	return checkAllowedKeys(raw, crushAllowedKeys, c.Name())
}
