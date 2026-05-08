package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchE2E_Backends(t *testing.T) {
	fake := buildFakeBackend(t)
	cases := []struct {
		name      string
		args      []string
		manifest  func(workspace, record, extra, resource1, resource2 string) string
		wantArgv  func(fake, extra, resource1, resource2 string) []string
		wantNoArg []string
	}{
		{
			name: "claude-code",
			args: []string{"launch", "claude-code", "--resume", "--add", "EXTRA", "--", "--passthru", "value"},
			manifest: func(workspace, record, extra, resource1, resource2 string) string {
				return fmt.Sprintf(`
name = "claude-code"
kind = "claude-code"
workspace = %q
resources = [%q, %q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "7"
AIH_MANIFEST_ENV = "manifest"

[backend.claude-code]
binary = %q
agent = "infra"
model = "sonnet"
permission_mode = "default"
`, workspace, resource1, resource2, record, fake)
			},
			wantArgv: func(fake, extra, resource1, resource2 string) []string {
				return []string{fake, "--continue", "--add-dir", resource1, "--add-dir", resource2, "--add-dir", extra, "--agent", "infra", "--model", "sonnet", "--permission-mode", "default", "--passthru", "value"}
			},
		},
		{
			name: "codex",
			args: []string{"launch", "codex", "--add", "EXTRA", "--", "--passthru", "value"},
			manifest: func(workspace, record, extra, resource1, resource2 string) string {
				return fmt.Sprintf(`
name = "codex"
kind = "codex"
workspace = %q
resources = [%q, %q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "7"
AIH_MANIFEST_ENV = "manifest"

[backend.codex]
binary = %q
model = "gpt-5"
sandbox = "read-only"
profile = "work"
ask_for_approval = "never"
full_auto = true
`, workspace, resource1, resource2, record, fake)
			},
			wantArgv: func(fake, extra, resource1, resource2 string) []string {
				return []string{fake, "--model", "gpt-5", "--sandbox", "read-only", "--profile", "work", "--ask-for-approval", "never", "--full-auto", "--add-dir", resource1, "--add-dir", resource2, "--add-dir", extra, "--passthru", "value"}
			},
		},
		{
			name: "kiro",
			args: []string{"launch", "kiro", "--add", "EXTRA", "--", "--passthru", "value"},
			manifest: func(workspace, record, extra, resource1, resource2 string) string {
				return fmt.Sprintf(`
name = "kiro"
kind = "kiro"
workspace = %q
resources = [%q, %q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "7"
AIH_MANIFEST_ENV = "manifest"

[backend.kiro]
binary = %q
mode = "spec"
profile = "work"
new_window = true
`, workspace, resource1, resource2, record, fake)
			},
			wantArgv: func(fake, extra, resource1, resource2 string) []string {
				return []string{fake, "chat", "--mode", "spec", "--profile", "work", "--new-window", "--add-file", resource1, "--add-file", resource2, "--add-file", extra, "--passthru", "value"}
			},
		},
		{
			name: "crush",
			args: []string{"launch", "crush", "--resume", "--", "--passthru", "value"},
			manifest: func(workspace, record, extra, resource1, resource2 string) string {
				return fmt.Sprintf(`
name = "crush"
kind = "crush"
workspace = %q
resources = [%q, %q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "7"
AIH_MANIFEST_ENV = "manifest"

[backend.crush]
binary = %q
data_dir = %q
yolo = true
debug = true
host = "127.0.0.1"
`, workspace, resource1, resource2, record, fake, filepath.Join(workspace, "crush-data"))
			},
			wantArgv: func(fake, extra, resource1, resource2 string) []string {
				return []string{fake, "--continue", "--data-dir", filepath.Join(filepath.Dir(extra), "workspace", "crush-data"), "--yolo", "--debug", "--host", "127.0.0.1", "--passthru", "value"}
			},
			wantNoArg: []string{"--add-dir", "--add-file"},
		},
		{
			name: "opencode",
			args: []string{"launch", "opencode", "--resume", "--", "--passthru", "value"},
			manifest: func(workspace, record, extra, resource1, resource2 string) string {
				return fmt.Sprintf(`
name = "opencode"
kind = "opencode"
workspace = %q
resources = [%q, %q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "7"
AIH_MANIFEST_ENV = "manifest"

[backend.opencode]
binary = %q
model = "claude"
agent = "infra"
log_level = "INFO"
pure = true
`, workspace, resource1, resource2, record, fake)
			},
			wantArgv: func(fake, extra, resource1, resource2 string) []string {
				return []string{fake, "--continue", "--model", "claude", "--agent", "infra", "--log-level", "INFO", "--pure", "--passthru", "value"}
			},
			wantNoArg: []string{"--add-dir", "--add-file"},
		},
		{
			name: "generic",
			args: []string{"launch", "generic", "--", "--passthru", "value"},
			manifest: func(workspace, record, extra, resource1, resource2 string) string {
				return fmt.Sprintf(`
name = "generic"
kind = "generic"
workspace = %q
command = [%q, "--base"]
resources = [%q, %q]

[env]
AIH_FAKE_RECORD = %q
AIH_FAKE_EXIT = "7"
AIH_MANIFEST_ENV = "manifest"
`, workspace, fake, resource1, resource2, record)
			},
			wantArgv: func(fake, extra, resource1, resource2 string) []string {
				return []string{fake, "--base", "--passthru", "value"}
			},
			wantNoArg: []string{"--add-dir", "--add-file"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupAihHome(t)
			parentValue := "parent-" + strings.ReplaceAll(tc.name, "-", "_")
			t.Setenv("AIH_PARENT_ENV", parentValue)
			dir := t.TempDir()
			workspace := filepath.Join(dir, "workspace")
			if err := os.MkdirAll(workspace, 0o755); err != nil {
				t.Fatal(err)
			}
			record := filepath.Join(dir, "record.json")
			extra := filepath.Join(dir, "extra")
			resource1 := filepath.Join(dir, "resource1")
			resource2 := filepath.Join(dir, "resource2")
			args := replaceArg(tc.args, "EXTRA", extra)
			writeManifest(t, tc.name, tc.manifest(workspace, record, extra, resource1, resource2))

			res := runAih(t, args...)
			if res.ExitCode != 7 {
				t.Fatalf("ExitCode = %d, want 7; stderr=%s err=%v", res.ExitCode, res.Stderr, res.Err)
			}
			rec := recordedRun(t, record)
			wantArgv := tc.wantArgv(fake, extra, resource1, resource2)
			assertSlice(t, rec.Argv, wantArgv)
			if !samePath(t, rec.Cwd, workspace) {
				t.Errorf("cwd = %q, want %q", rec.Cwd, workspace)
			}
			if rec.Env["AIH_PARENT_ENV"] != parentValue {
				t.Errorf("AIH_PARENT_ENV = %q, want %q", rec.Env["AIH_PARENT_ENV"], parentValue)
			}
			if rec.Env["AIH_MANIFEST_ENV"] != "manifest" {
				t.Errorf("AIH_MANIFEST_ENV = %q, want manifest", rec.Env["AIH_MANIFEST_ENV"])
			}
			for _, arg := range tc.wantNoArg {
				assertNotContains(t, rec.Argv, arg)
			}
		})
	}
}

func replaceArg(args []string, old, new string) []string {
	out := append([]string(nil), args...)
	for i, arg := range out {
		if arg == old {
			out[i] = new
		}
	}
	return out
}

func assertSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("argv len = %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("argv[%d] = %q, want %q\ngot:  %v\nwant: %v", i, got[i], want[i], got, want)
		}
	}
}

func samePath(t *testing.T, got, want string) bool {
	t.Helper()
	gotEval, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatal(err)
	}
	wantEval, err := filepath.EvalSymlinks(want)
	if err != nil {
		t.Fatal(err)
	}
	return gotEval == wantEval
}

func assertNotContains(t *testing.T, got []string, value string) {
	t.Helper()
	for _, item := range got {
		if item == value {
			t.Fatalf("argv unexpectedly contains %q: %v", value, got)
		}
	}
}
