package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestLaunch_EchoExitsZero(t *testing.T) {
	dir := t.TempDir()
	m := &manifest.Manifest{Name: "echo-test", Workspace: dir}
	res, err := Launch(m, []string{"true"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
}

func TestLaunch_PropagatesNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	m := &manifest.Manifest{Name: "false-test", Workspace: dir}
	res, err := Launch(m, []string{"false"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if res.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", res.ExitCode)
	}
}

func TestLaunch_RunsInWorkspace(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")
	m := &manifest.Manifest{Name: "ws-test", Workspace: dir}
	// `touch marker.txt` (relative) — only succeeds if cwd was set.
	res, err := Launch(m, []string{"touch", "marker.txt"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d", res.ExitCode)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("marker not created in workspace: %v", err)
	}
}

func TestLaunch_AppliesEnv(t *testing.T) {
	dir := t.TempDir()
	m := &manifest.Manifest{
		Name:      "env-test",
		Workspace: dir,
		Env:       map[string]string{"AIH_TEST_VAR": "hello-aih"},
	}
	out := filepath.Join(dir, "env.txt")
	// sh -c writes $AIH_TEST_VAR to a file in the workspace.
	res, err := Launch(m, []string{"sh", "-c", "printf %s \"$AIH_TEST_VAR\" > env.txt"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d", res.ExitCode)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello-aih" {
		t.Errorf("env not applied: file contents = %q", got)
	}
}

func TestLaunch_RejectsMissingWorkspace(t *testing.T) {
	m := &manifest.Manifest{Name: "x", Workspace: "/nonexistent/path/" + t.Name()}
	if _, err := Launch(m, []string{"true"}); err == nil {
		t.Fatal("expected error for missing workspace")
	}
}

func TestLaunch_RejectsEmptyArgv(t *testing.T) {
	m := &manifest.Manifest{Name: "x", Workspace: t.TempDir()}
	if _, err := Launch(m, nil); err == nil {
		t.Fatal("expected error for empty argv")
	}
}

func TestMergeEnv_ManifestOverrides(t *testing.T) {
	base := []string{"FOO=base", "BAR=base"}
	got := mergeEnv(base, map[string]string{"FOO": "override", "BAZ": "new"})
	want := map[string]string{"FOO": "override", "BAR": "base", "BAZ": "new"}
	for _, kv := range got {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				k, v := kv[:i], kv[i+1:]
				if want[k] != v {
					t.Errorf("env[%s] = %q, want %q", k, v, want[k])
				}
				delete(want, k)
				break
			}
		}
	}
	if len(want) != 0 {
		t.Errorf("missing env keys: %v", want)
	}
}
