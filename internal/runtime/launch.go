// Package runtime executes a backend command in the manifest-declared
// workspace with the resolved env. Signals are forwarded to the child so
// Ctrl-C terminates cleanly and the parent returns the child's exit code.
package runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/vaughngit/aih/internal/manifest"
)

// Result describes a completed launch. ExitCode mirrors the child's exit
// status; Signal is non-empty when the child was terminated by a signal.
type Result struct {
	ExitCode int
	Signal   string
}

// Launch executes argv in the manifest's workspace with the inherited env
// extended by the manifest's [env] table. stdio is wired to the parent
// terminal. SIGINT/SIGTERM/SIGHUP are forwarded to the child.
func Launch(m *manifest.Manifest, argv []string) (*Result, error) {
	if len(argv) == 0 {
		return nil, errors.New("runtime: empty argv")
	}
	if m.Workspace == "" {
		return nil, errors.New("runtime: manifest has no workspace")
	}
	if info, err := os.Stat(m.Workspace); err != nil {
		return nil, fmt.Errorf("runtime: workspace %s: %w", m.Workspace, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("runtime: workspace %s is not a directory", m.Workspace)
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = m.Workspace
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = mergeEnv(os.Environ(), m.Env)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("runtime: start %s: %w", argv[0], err)
	}

	stop := forwardSignals(cmd.Process)
	defer stop()

	err := cmd.Wait()
	res := &Result{}
	if err == nil {
		res.ExitCode = 0
		return res, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		if sig, ok := signalName(exitErr); ok {
			res.Signal = sig
		}
		return res, nil
	}
	return nil, fmt.Errorf("runtime: wait: %w", err)
}

// mergeEnv produces a child env vector. The base (parent) env is preserved;
// manifest entries override or extend. Output is sorted for deterministic
// behavior in tests.
func mergeEnv(base []string, override map[string]string) []string {
	merged := map[string]string{}
	for _, kv := range base {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				merged[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	for k, v := range override {
		merged[k] = v
	}
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+merged[k])
	}
	return out
}
