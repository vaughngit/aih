package backends

import (
	"github.com/vaughngit/aih/internal/manifest"
)

func init() { Register(&ClaudeCode{}) }

// ClaudeCode wraps the `claude` CLI (Claude Code).
//
// Resume semantics: aih's --resume is mapped to `claude --continue`, which
// resumes the most recent session in the current cwd without an
// interactive picker. claude's own `--resume` flag opens a picker, which
// is unhelpful when scripted from aih.
//
// Resources and --add: passed through as repeated --add-dir flags.
type ClaudeCode struct{}

func (*ClaudeCode) Name() string             { return "claude-code" }
func (*ClaudeCode) SupportsResume() bool     { return true }
func (*ClaudeCode) SupportsAddDir() bool     { return true }
func (*ClaudeCode) DefaultLogSubdir() string { return "claude-code" }

var claudeCodeAllowedKeys = map[string]keyType{
	"binary":               keyString,
	"agent":                keyString,
	"model":                keyString,
	"permission_mode":      keyString,
	"append_system_prompt": keyString,
}

var claudeCodePermissionModes = []string{
	"acceptEdits", "auto", "bypassPermissions", "default", "dontAsk", "plan",
}

func (cc *ClaudeCode) BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error) {
	cfg := m.BackendConfig(cc.Name())
	argv := []string{stringOrDefault(cfg, "binary", "claude")}

	if opts.Resume {
		argv = append(argv, "--continue")
	}
	for _, r := range m.Resources {
		argv = append(argv, "--add-dir", r)
	}
	for _, d := range opts.AddDirs {
		argv = append(argv, "--add-dir", d)
	}
	if v := stringOrDefault(cfg, "agent", ""); v != "" {
		argv = append(argv, "--agent", v)
	}
	if v := stringOrDefault(cfg, "model", ""); v != "" {
		argv = append(argv, "--model", v)
	}
	if v := stringOrDefault(cfg, "permission_mode", ""); v != "" {
		argv = append(argv, "--permission-mode", v)
	}
	if v := stringOrDefault(cfg, "append_system_prompt", ""); v != "" {
		argv = append(argv, "--append-system-prompt", v)
	}

	argv = append(argv, opts.PassThru...)
	return argv, nil
}

func (cc *ClaudeCode) ValidateConfig(raw map[string]any) error {
	if err := checkAllowedKeys(raw, claudeCodeAllowedKeys, cc.Name()); err != nil {
		return err
	}
	return checkEnum(stringOrDefault(raw, "permission_mode", ""),
		"permission_mode", cc.Name(), claudeCodePermissionModes)
}
