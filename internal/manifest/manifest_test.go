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
