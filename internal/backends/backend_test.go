package backends

import (
	"strings"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestGetGeneric(t *testing.T) {
	b, err := Get("generic")
	if err != nil {
		t.Fatal(err)
	}
	if b.Name() != "generic" {
		t.Errorf("Name = %q", b.Name())
	}
	if b.SupportsResume() || b.SupportsAddDir() {
		t.Errorf("generic should not support resume or add-dir")
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("does-not-exist")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "have:") {
		t.Errorf("error should list registered backends, got: %v", err)
	}
}

func TestGenericBuildCommand_Verbatim(t *testing.T) {
	b, _ := Get("generic")
	m := &manifest.Manifest{Name: "x", Command: []string{"echo", "hello"}}
	got, err := b.BuildCommand(m, LaunchOpts{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"echo", "hello"}
	if !equalSlice(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGenericBuildCommand_PassThrough(t *testing.T) {
	b, _ := Get("generic")
	m := &manifest.Manifest{Name: "x", Command: []string{"echo"}}
	got, err := b.BuildCommand(m, LaunchOpts{PassThru: []string{"--flag", "value"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"echo", "--flag", "value"}
	if !equalSlice(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGenericBuildCommand_RejectsResume(t *testing.T) {
	b, _ := Get("generic")
	m := &manifest.Manifest{Name: "x", Command: []string{"echo"}}
	if _, err := b.BuildCommand(m, LaunchOpts{Resume: true}); err == nil {
		t.Fatal("expected error for --resume on generic")
	}
}

func TestGenericBuildCommand_EmptyCommandErrors(t *testing.T) {
	b, _ := Get("generic")
	m := &manifest.Manifest{Name: "x"}
	if _, err := b.BuildCommand(m, LaunchOpts{}); err == nil {
		t.Fatal("expected error for empty Command")
	}
}

func TestGenericValidateConfig_RejectsKeys(t *testing.T) {
	b, _ := Get("generic")
	if err := b.ValidateConfig(nil); err != nil {
		t.Errorf("nil cfg should be ok, got %v", err)
	}
	if err := b.ValidateConfig(map[string]any{"agent": "x"}); err == nil {
		t.Fatal("expected error for non-empty generic config")
	}
}

func TestNames_IncludesGeneric(t *testing.T) {
	names := Names()
	found := false
	for _, n := range names {
		if n == "generic" {
			found = true
		}
	}
	if !found {
		t.Errorf("Names() = %v, missing 'generic'", names)
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
