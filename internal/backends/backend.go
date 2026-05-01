// Package backends defines the Backend plugin interface and a process-wide
// registry. Phase 1 ships only the generic backend; Phase 2 adds claude-code,
// codex, crush, kiro, opencode.
package backends

import (
	"fmt"
	"sort"
	"sync"

	"github.com/vaughngit/aih/internal/manifest"
)

// LaunchOpts carries per-invocation toggles from CLI flags.
type LaunchOpts struct {
	Resume    bool
	AddDirs   []string
	PassThru  []string
	BackendID string // optional override; empty = use manifest.Kind
}

// Backend is the per-CLI plugin contract. New backends register themselves
// in their package's init() via Register.
type Backend interface {
	// Name is the discriminator value (matches manifest.Kind).
	Name() string

	// BuildCommand returns the full argv vector (binary + args) used to launch.
	BuildCommand(m *manifest.Manifest, opts LaunchOpts) ([]string, error)

	SupportsResume() bool
	SupportsAddDir() bool

	// DefaultLogSubdir is the directory name under the per-OS log root.
	DefaultLogSubdir() string

	// ValidateConfig inspects the [backend.<Name>] table contents.
	ValidateConfig(raw map[string]any) error
}

var (
	mu       sync.RWMutex
	registry = map[string]Backend{}
)

// Register adds a backend to the registry. Duplicate names panic — that's a
// programming error, not a runtime condition.
func Register(b Backend) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[b.Name()]; exists {
		panic(fmt.Sprintf("backends: duplicate Register for %q", b.Name()))
	}
	registry[b.Name()] = b
}

// Get returns the backend matching name, or an error listing what is
// registered. Callers pass manifest.Kind (or the --backend override).
func Get(name string) (Backend, error) {
	mu.RLock()
	defer mu.RUnlock()
	b, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("backends: unknown backend %q (have: %v)", name, sortedNames())
	}
	return b, nil
}

// Names returns all registered backend names, sorted.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	return sortedNames()
}

func sortedNames() []string {
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
