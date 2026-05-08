package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchE2E_BackendOverrideUsesOverrideConfig(t *testing.T) {
	setupAihHome(t)
	fake := buildFakeBackend(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(dir, "record.json")
	writeManifest(t, "override", fmt.Sprintf(`
name = "override"
kind = "claude-code"
workspace = %q
command = [%q, "--generic"]

[env]
AIH_FAKE_RECORD = %q

[backend.claude-code]
binary = "/does/not/run"
`, workspace, fake, record))

	res := runAih(t, "launch", "override", "--backend", "generic", "--", "--passthru")
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("launch failed: exit=%d err=%v stderr=%s", res.ExitCode, res.Err, res.Stderr)
	}
	rec := recordedRun(t, record)
	assertSlice(t, rec.Argv, []string{fake, "--generic", "--passthru"})
}

func TestLaunchE2E_CodexResumeArgv(t *testing.T) {
	setupAihHome(t)
	fake := buildFakeBackend(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(dir, "record.json")
	writeManifest(t, "codex-resume", fmt.Sprintf(`
name = "codex-resume"
kind = "codex"
workspace = %q

[env]
AIH_FAKE_RECORD = %q

[backend.codex]
binary = %q
model = "ignored-on-resume"
`, workspace, record, fake))

	res := runAih(t, "launch", "codex-resume", "--resume", "--", "--passthru")
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("launch failed: exit=%d err=%v stderr=%s", res.ExitCode, res.Err, res.Stderr)
	}
	rec := recordedRun(t, record)
	assertSlice(t, rec.Argv, []string{fake, "resume", "--last", "--passthru"})
}

func TestLaunchE2E_RejectsUnsupportedResume(t *testing.T) {
	setupAihHome(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, "kiro-no-resume", fmt.Sprintf(`
name = "kiro-no-resume"
kind = "kiro"
workspace = %q
`, workspace))

	res := runAih(t, "launch", "kiro-no-resume", "--resume")
	if res.ExitCode != 1 || res.Err == nil {
		t.Fatalf("ExitCode=%d Err=%v, want command error", res.ExitCode, res.Err)
	}
	if !strings.Contains(res.Stderr, `backend "kiro" does not support --resume`) {
		t.Fatalf("stderr = %q", res.Stderr)
	}
}

func TestLaunchE2E_RejectsUnsupportedAdd(t *testing.T) {
	setupAihHome(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, "crush-no-add", fmt.Sprintf(`
name = "crush-no-add"
kind = "crush"
workspace = %q
`, workspace))

	res := runAih(t, "launch", "crush-no-add", "--add", filepath.Join(dir, "extra"))
	if res.ExitCode != 1 || res.Err == nil {
		t.Fatalf("ExitCode=%d Err=%v, want command error", res.ExitCode, res.Err)
	}
	if !strings.Contains(res.Stderr, `backend "crush" does not support --add`) {
		t.Fatalf("stderr = %q", res.Stderr)
	}
}

func TestLaunchE2E_CodexResumeRejectsAdd(t *testing.T) {
	setupAihHome(t)
	fake := buildFakeBackend(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, "codex-resume-add", fmt.Sprintf(`
name = "codex-resume-add"
kind = "codex"
workspace = %q

[backend.codex]
binary = %q
`, workspace, fake))

	res := runAih(t, "launch", "codex-resume-add", "--resume", "--add", filepath.Join(dir, "extra"))
	if res.ExitCode != 1 || res.Err == nil {
		t.Fatalf("ExitCode=%d Err=%v, want command error", res.ExitCode, res.Err)
	}
	if !strings.Contains(res.Stderr, "codex: --add cannot be combined with --resume") {
		t.Fatalf("stderr = %q", res.Stderr)
	}
}

func TestLaunchE2E_ValidatesBackendConfig(t *testing.T) {
	setupAihHome(t)
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, "bad-config", fmt.Sprintf(`
name = "bad-config"
kind = "opencode"
workspace = %q

[backend.opencode]
log_level = "TRACE"
`, workspace))

	res := runAih(t, "launch", "bad-config")
	if res.ExitCode != 1 || res.Err == nil {
		t.Fatalf("ExitCode=%d Err=%v, want validation error", res.ExitCode, res.Err)
	}
	if !strings.Contains(res.Stderr, "log_level") || !strings.Contains(res.Stderr, "TRACE") {
		t.Fatalf("stderr = %q", res.Stderr)
	}
}
