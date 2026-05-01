// Package manifest defines the on-disk TOML schema for aih agent manifests
// and provides loading + path expansion. Validation lives here for the core
// schema; per-backend [backend.<name>] tables are dispatched to the matching
// backend plugin (see internal/backends).
package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Manifest mirrors the TOML schema. The Kind field is the backend
// discriminator (renamed from "backend" in the spec because TOML disallows a
// key being both a value and a table-prefix). Per-backend config lives in
// BackendCfg, populated from the [backend.<name>] tables.
type Manifest struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Kind        string            `toml:"kind"`
	Workspace   string            `toml:"workspace"`
	Resources   []string          `toml:"resources"`
	Command     []string          `toml:"command"`
	Env         map[string]string `toml:"env"`
	Hooks       Hooks             `toml:"hooks"`
	BackendCfg  map[string]any    `toml:"backend"`

	// Source records where the manifest was loaded from. Not part of TOML.
	Source string `toml:"-"`
}

// Hooks is the [hooks] table. Phase 1 stub; Phase 4 wires this to a runner.
type Hooks struct {
	PreLaunch  []string `toml:"pre_launch"`
	PostLaunch []string `toml:"post_launch"`
}

// Load reads a manifest from disk, decodes the TOML, and expands `~` and
// $VAR references in path-like fields.
func Load(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest: read %s: %w", path, err)
	}
	m := &Manifest{Source: path}
	dec := toml.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(m); err != nil {
		var miss *toml.StrictMissingError
		if errors.As(err, &miss) {
			return nil, fmt.Errorf("manifest: parse %s: unknown field(s):\n%s", path, miss.String())
		}
		return nil, fmt.Errorf("manifest: parse %s: %w", path, err)
	}
	if err := m.expand(); err != nil {
		return nil, err
	}
	return m, nil
}

// expand applies tilde and env-var expansion to path-like fields.
func (m *Manifest) expand() error {
	w, err := ExpandPath(m.Workspace)
	if err != nil {
		return fmt.Errorf("manifest: expand workspace: %w", err)
	}
	m.Workspace = w

	for i, r := range m.Resources {
		exp, err := ExpandPath(r)
		if err != nil {
			return fmt.Errorf("manifest: expand resource %q: %w", r, err)
		}
		m.Resources[i] = exp
	}

	for k, v := range m.Env {
		m.Env[k] = os.ExpandEnv(expandTilde(v))
	}
	return nil
}

// ExpandPath resolves `~` and `$VAR` / `${VAR}` references in a path.
// An empty input is returned unchanged.
func ExpandPath(p string) (string, error) {
	if p == "" {
		return "", nil
	}
	p = expandTilde(p)
	p = os.ExpandEnv(p)
	if !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}

func expandTilde(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}
