package backends

import (
	"github.com/vaughngit/aih/internal/manifest"
)

func init() { Register(&OpenCode{}) }

// OpenCode wraps the `opencode` TUI.
//
// Resume semantics: aih's --resume becomes `opencode --continue` (last
// session). For a specific session id use passthru: `aih launch x -- --session <id>`.
//
// No add-dir support. opencode's only context-injection mechanism is the
// `[project]` positional, which cmd.Dir already handles. `manifest.Resources`
// is silently ignored.
type OpenCode struct{}

func (*OpenCode) Name() string             { return "opencode" }
func (*OpenCode) SupportsResume() bool     { return true }
func (*OpenCode) SupportsAddDir() bool     { return false }
func (*OpenCode) DefaultLogSubdir() string { return "opencode" }

var opencodeAllowedKeys = map[string]keyType{
	"binary":    keyString,
	"model":     keyString,
	"agent":     keyString,
	"log_level": keyString,
	"pure":      keyBool,
}

var opencodeLogLevels = []string{"DEBUG", "INFO", "WARN", "ERROR"}

func (oc *OpenCode) BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error) {
	cfg := m.BackendConfig(oc.Name())
	argv := []string{stringOrDefault(cfg, "binary", "opencode")}

	if opts.Resume {
		argv = append(argv, "--continue")
	}
	if v := stringOrDefault(cfg, "model", ""); v != "" {
		argv = append(argv, "--model", v)
	}
	if v := stringOrDefault(cfg, "agent", ""); v != "" {
		argv = append(argv, "--agent", v)
	}
	if v := stringOrDefault(cfg, "log_level", ""); v != "" {
		argv = append(argv, "--log-level", v)
	}
	if boolOrDefault(cfg, "pure", false) {
		argv = append(argv, "--pure")
	}

	argv = append(argv, opts.PassThru...)
	return argv, nil
}

func (oc *OpenCode) ValidateConfig(raw map[string]any) error {
	if err := checkAllowedKeys(raw, opencodeAllowedKeys, oc.Name()); err != nil {
		return err
	}
	return checkEnum(stringOrDefault(raw, "log_level", ""),
		"log_level", oc.Name(), opencodeLogLevels)
}
