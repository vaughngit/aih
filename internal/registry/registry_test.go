package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func writeManifest(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".toml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func setAIHHome(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("AIH_HOME", root)
	return filepath.Join(root, "agents")
}

func TestLoadCentral_Found(t *testing.T) {
	dir := setAIHHome(t)
	writeManifest(t, dir, "alpha", `
name = "alpha"
kind = "generic"
workspace = "/tmp"
`)
	m, err := LoadCentral("alpha")
	if err != nil {
		t.Fatalf("LoadCentral: %v", err)
	}
	if m.Name != "alpha" || m.Kind != "generic" {
		t.Errorf("got %+v", m)
	}
}

func TestLoadCentral_NotFound(t *testing.T) {
	setAIHHome(t)
	if _, err := LoadCentral("missing"); err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestLoadCentral_RejectsTraversal(t *testing.T) {
	setAIHHome(t)
	for _, bad := range []string{"", "../etc/passwd", "a/b", "..", "a\\b"} {
		if _, err := LoadCentral(bad); err == nil {
			t.Errorf("LoadCentral(%q): expected error, got nil", bad)
		}
	}
}

func TestList_SortedAndSkipsBroken(t *testing.T) {
	dir := setAIHHome(t)
	writeManifest(t, dir, "zulu", `
name = "zulu"
kind = "generic"
workspace = "/tmp"
`)
	writeManifest(t, dir, "alpha", `
name = "alpha"
kind = "generic"
workspace = "/tmp"
`)
	// Broken file — must be skipped, not error out.
	writeManifest(t, dir, "broken", "this is = not valid TOML[[[")
	// Non-toml file in the dir should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2 (good ones); got names: %v", len(got), names(got))
	}
	if got[0].Name != "alpha" || got[1].Name != "zulu" {
		t.Errorf("not sorted: %v", names(got))
	}
}

func names(ms []*manifest.Manifest) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.Name
	}
	return out
}

func TestList_MissingDirIsEmpty(t *testing.T) {
	t.Setenv("AIH_HOME", filepath.Join(t.TempDir(), "does-not-exist"))
	got, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

