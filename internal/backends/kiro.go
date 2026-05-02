package backends

import (
	"github.com/vaughngit/aih/internal/manifest"
)

func init() { Register(&Kiro{}) }

// Kiro wraps `kiro chat`, the agent entry point of the Kiro editor.
//
// No resume: `kiro chat` has no resume flag; sessions live in the editor
// window's state, not on the CLI.
//
// Resources / --add: passed as repeated --add-file flags. Kiro's flag is
// file-only — directory paths will be rejected by kiro itself, not here.
// The manifest author chooses what to put in `resources`.
type Kiro struct{}

func (*Kiro) Name() string             { return "kiro" }
func (*Kiro) SupportsResume() bool     { return false }
func (*Kiro) SupportsAddDir() bool     { return true }
func (*Kiro) DefaultLogSubdir() string { return "kiro" }

var kiroAllowedKeys = map[string]keyType{
	"binary":       keyString,
	"mode":         keyString,
	"profile":      keyString,
	"new_window":   keyBool,
	"reuse_window": keyBool,
	"maximize":     keyBool,
}

func (k *Kiro) BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error) {
	cfg := m.BackendConfig(k.Name())
	binary := stringOrDefault(cfg, "binary", "kiro")
	argv := []string{binary, "chat"}

	if v := stringOrDefault(cfg, "mode", ""); v != "" {
		argv = append(argv, "--mode", v)
	}
	if v := stringOrDefault(cfg, "profile", ""); v != "" {
		argv = append(argv, "--profile", v)
	}
	if boolOrDefault(cfg, "new_window", false) {
		argv = append(argv, "--new-window")
	}
	if boolOrDefault(cfg, "reuse_window", false) {
		argv = append(argv, "--reuse-window")
	}
	if boolOrDefault(cfg, "maximize", false) {
		argv = append(argv, "--maximize")
	}
	for _, r := range m.Resources {
		argv = append(argv, "--add-file", r)
	}
	for _, d := range opts.AddDirs {
		argv = append(argv, "--add-file", d)
	}

	argv = append(argv, opts.PassThru...)
	return argv, nil
}

func (k *Kiro) ValidateConfig(raw map[string]any) error {
	return checkAllowedKeys(raw, kiroAllowedKeys, k.Name())
}
