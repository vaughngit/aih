package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_Minimal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	src := `
name = "test"
description = "smoke"
kind = "generic"
workspace = "` + dir + `"
command = ["echo", "hello"]
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Name != "test" {
		t.Errorf("Name = %q, want %q", m.Name, "test")
	}
	if m.Kind != "generic" {
		t.Errorf("Kind = %q, want %q", m.Kind, "generic")
	}
	if got, want := m.Command, []string{"echo", "hello"}; !equalSlice(got, want) {
		t.Errorf("Command = %v, want %v", got, want)
	}
	if m.Source != path {
		t.Errorf("Source = %q, want %q", m.Source, path)
	}
}

func TestLoad_KindBackendCoexist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kc.toml")
	src := `
name = "kc"
kind = "claude-code"
workspace = "` + dir + `"

[backend.claude-code]
agent = "kubernetes"
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Kind != "claude-code" {
		t.Errorf("Kind = %q", m.Kind)
	}
	cc, ok := m.BackendCfg["claude-code"].(map[string]any)
	if !ok {
		t.Fatalf("BackendCfg[claude-code] missing or wrong type: %#v", m.BackendCfg)
	}
	if cc["agent"] != "kubernetes" {
		t.Errorf("backend.claude-code.agent = %v, want kubernetes", cc["agent"])
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no HOME: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "tilde.toml")
	src := `
name = "tilde"
kind = "generic"
workspace = "~"
resources = ["~/sub", "~"]
[env]
P = "~/path"
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Workspace != home {
		t.Errorf("Workspace = %q, want %q", m.Workspace, home)
	}
	if got := m.Resources[0]; got != filepath.Join(home, "sub") {
		t.Errorf("Resources[0] = %q", got)
	}
	if got := m.Resources[1]; got != home {
		t.Errorf("Resources[1] = %q", got)
	}
	if got := m.Env["P"]; got != filepath.Join(home, "path") {
		t.Errorf("Env[P] = %q", got)
	}
}

func TestLoad_EnvExpansion(t *testing.T) {
	t.Setenv("AIH_TEST_WS", "/tmp/aih-ws-fixture")
	dir := t.TempDir()
	path := filepath.Join(dir, "envvar.toml")
	src := `
name = "envvar"
kind = "generic"
workspace = "$AIH_TEST_WS"
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Workspace != "/tmp/aih-ws-fixture" {
		t.Errorf("Workspace = %q, want /tmp/aih-ws-fixture", m.Workspace)
	}
}

func TestLoad_RejectsUnknownTopLevelKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	src := `
name = "x"
kind = "generic"
workspace = "/tmp"
unknown_typo_field = "oops"
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for unknown field, got nil")
	} else if !strings.Contains(err.Error(), "unknown_typo_field") {
		t.Errorf("error should mention bad field, got: %v", err)
	}
}

func TestLoad_RejectsBackendStringTableCollision(t *testing.T) {
	// The original spec example. Must remain rejected — confirms the
	// `kind` rename is load-bearing and that we don't silently accept
	// the broken layout.
	dir := t.TempDir()
	path := filepath.Join(dir, "collision.toml")
	src := `
name = "x"
backend = "claude-code"

[backend.claude-code]
agent = "kubernetes"
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected TOML parse error for backend string+table collision, got nil")
	}
}

func TestValidate_AllFieldsPresent(t *testing.T) {
	m := &Manifest{Name: "x", Kind: "generic", Workspace: "/tmp", Source: "x.toml"}
	if err := m.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidate_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		m    *Manifest
		want string // substring expected in the error
	}{
		{"no name", &Manifest{Kind: "generic", Workspace: "/tmp"}, "Name"},
		{"no kind", &Manifest{Name: "x", Workspace: "/tmp"}, "Kind"},
		{"no workspace", &Manifest{Name: "x", Kind: "generic"}, "Workspace"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.m.Validate()
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestBackendConfig_ExtractsTable(t *testing.T) {
	m := &Manifest{
		BackendCfg: map[string]any{
			"claude-code": map[string]any{"agent": "k8s", "model": "opus"},
		},
	}
	cfg := m.BackendConfig("claude-code")
	if cfg["agent"] != "k8s" {
		t.Errorf("agent = %v", cfg["agent"])
	}
	if got := m.BackendConfig("missing"); got != nil {
		t.Errorf("missing should be nil, got %v", got)
	}
	if got := (&Manifest{}).BackendConfig("any"); got != nil {
		t.Errorf("nil BackendCfg should yield nil, got %v", got)
	}
}

func TestExpandPath_Empty(t *testing.T) {
	got, err := ExpandPath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("ExpandPath(\"\") = %q, want \"\"", got)
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
