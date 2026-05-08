package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLaunchPropagatesChildExitCode(t *testing.T) {
	setupAihHome(t)
	fake := buildFakeBackend(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(dir, "record.json")
	writeManifest(t, "exit-code", fmt.Sprintf(`
name = "exit-code"
kind = "generic"
workspace = %q
command = [%q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "42"
`, workspace, fake, record))

	res := runAih(t, "launch", "exit-code")
	if res.ExitCode != 42 {
		t.Fatalf("ExitCode = %d, want 42; stderr=%s err=%v", res.ExitCode, res.Stderr, res.Err)
	}
	rec := recordedRun(t, record)
	assertSlice(t, rec.Argv, []string{fake})
}
