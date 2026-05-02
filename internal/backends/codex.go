package backends

import (
	"fmt"

	"github.com/vaughngit/aih/internal/manifest"
)

func init() { Register(&Codex{}) }

// Codex wraps the OpenAI `codex` CLI.
//
// Resume semantics: aih's --resume becomes `codex resume --last`, which
// continues the most recent recorded session without showing the picker.
// In resume mode argv is just `codex resume --last [passthru]` — config
// flags and add-dirs only apply to new sessions and are rejected here so
// the failure mode is loud rather than silently ignored.
type Codex struct{}

func (*Codex) Name() string             { return "codex" }
func (*Codex) SupportsResume() bool     { return true }
func (*Codex) SupportsAddDir() bool     { return true }
func (*Codex) DefaultLogSubdir() string { return "codex" }

var codexAllowedKeys = map[string]keyType{
	"binary":           keyString,
	"model":            keyString,
	"sandbox":          keyString,
	"profile":          keyString,
	"ask_for_approval": keyString,
	"full_auto":        keyBool,
}

var (
	codexSandboxes = []string{"read-only", "workspace-write", "danger-full-access"}
	codexApprovals = []string{"untrusted", "on-failure", "on-request", "never"}
)

func (cx *Codex) BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error) {
	cfg := m.BackendConfig(cx.Name())
	binary := stringOrDefault(cfg, "binary", "codex")

	if opts.Resume {
		if len(opts.AddDirs) > 0 {
			return nil, fmt.Errorf("codex: --add cannot be combined with --resume (resume re-uses the original session's directories)")
		}
		argv := []string{binary, "resume", "--last"}
		argv = append(argv, opts.PassThru...)
		return argv, nil
	}

	argv := []string{binary}
	if v := stringOrDefault(cfg, "model", ""); v != "" {
		argv = append(argv, "--model", v)
	}
	if v := stringOrDefault(cfg, "sandbox", ""); v != "" {
		argv = append(argv, "--sandbox", v)
	}
	if v := stringOrDefault(cfg, "profile", ""); v != "" {
		argv = append(argv, "--profile", v)
	}
	if v := stringOrDefault(cfg, "ask_for_approval", ""); v != "" {
		argv = append(argv, "--ask-for-approval", v)
	}
	if boolOrDefault(cfg, "full_auto", false) {
		argv = append(argv, "--full-auto")
	}
	for _, r := range m.Resources {
		argv = append(argv, "--add-dir", r)
	}
	for _, d := range opts.AddDirs {
		argv = append(argv, "--add-dir", d)
	}

	argv = append(argv, opts.PassThru...)
	return argv, nil
}

func (cx *Codex) ValidateConfig(raw map[string]any) error {
	if err := checkAllowedKeys(raw, codexAllowedKeys, cx.Name()); err != nil {
		return err
	}
	if err := checkEnum(stringOrDefault(raw, "sandbox", ""),
		"sandbox", cx.Name(), codexSandboxes); err != nil {
		return err
	}
	return checkEnum(stringOrDefault(raw, "ask_for_approval", ""),
		"ask_for_approval", cx.Name(), codexApprovals)
}
