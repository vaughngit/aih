// Package registry resolves agent names to manifest files. v1 reads only the
// central registry at $AIH_HOME (default ~/.aih/agents/); Phase 3 adds
// project-local discovery and per-field merge.
package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vaughngit/aih/internal/manifest"
)

// CentralDir returns the directory holding central-registry manifests. Honors
// the AIH_HOME env var (test fixtures + future per-machine overrides); falls
// back to ~/.aih.
func CentralDir() (string, error) {
	if h := os.Getenv("AIH_HOME"); h != "" {
		return filepath.Join(h, "agents"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("registry: home dir: %w", err)
	}
	return filepath.Join(home, ".aih", "agents"), nil
}

// LoadCentral reads a single manifest from the central registry by name.
func LoadCentral(name string) (*manifest.Manifest, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	dir, err := CentralDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, name+".toml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("registry: no manifest %q at %s", name, path)
		}
		return nil, fmt.Errorf("registry: stat %s: %w", path, err)
	}
	return manifest.Load(path)
}

// List enumerates every manifest in the central registry, sorted by name.
// Manifests that fail to load are skipped silently for `aih list`; callers
// that want strict behavior can filter or call LoadCentral per-name.
func List() ([]*manifest.Manifest, error) {
	dir, err := CentralDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("registry: read %s: %w", dir, err)
	}
	out := make([]*manifest.Manifest, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		m, err := manifest.Load(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// validateName rejects names that would escape the central directory or
// otherwise produce surprising paths.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("registry: empty agent name")
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return fmt.Errorf("registry: invalid agent name %q", name)
	}
	return nil
}
