package backends

import (
	"strings"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestCodex_Registered(t *testing.T) {
	b, err := Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	if !b.SupportsResume() || !b.SupportsAddDir() {
		t.Errorf("codex should support resume and add-dir")
	}
}

func TestCodex_BuildCommand(t *testing.T) {
	cases := []struct {
		name string
		m    *manifest.Manifest
		opts LaunchOpts
		want []string
	}{
		{
			"bare",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{},
			[]string{"codex"},
		},
		{
			"resume_uses_subcommand",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{Resume: true},
			[]string{"codex", "resume", "--last"},
		},
		{
			"resume_passthru_after_subcommand",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{Resume: true, PassThru: []string{"keep going"}},
			[]string{"codex", "resume", "--last", "keep going"},
		},
		{
			"all_config_keys",
			&manifest.Manifest{
				Name: "x",
				BackendCfg: map[string]any{
					"codex": map[string]any{
						"binary":           "/opt/homebrew/bin/codex",
						"model":            "gpt-5-codex",
						"sandbox":          "workspace-write",
						"profile":          "work",
						"ask_for_approval": "on-request",
						"full_auto":        true,
					},
				},
			},
			LaunchOpts{},
			[]string{
				"/opt/homebrew/bin/codex",
				"--model", "gpt-5-codex",
				"--sandbox", "workspace-write",
				"--profile", "work",
				"--ask-for-approval", "on-request",
				"--full-auto",
			},
		},
		{
			"resources_become_add_dir",
			&manifest.Manifest{Name: "x", Resources: []string{"/a", "/b"}},
			LaunchOpts{AddDirs: []string{"/c"}},
			[]string{"codex", "--add-dir", "/a", "--add-dir", "/b", "--add-dir", "/c"},
		},
	}

	cx := &Codex{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cx.BuildCommand(tc.m, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !equalSlice(got, tc.want) {
				t.Errorf("got %v\nwant %v", got, tc.want)
			}
		})
	}
}

func TestCodex_ResumeRejectsAddDir(t *testing.T) {
	cx := &Codex{}
	_, err := cx.BuildCommand(&manifest.Manifest{Name: "x"},
		LaunchOpts{Resume: true, AddDirs: []string{"/x"}})
	if err == nil || !strings.Contains(err.Error(), "--add cannot be combined with --resume") {
		t.Fatalf("expected resume+add error, got %v", err)
	}
}

func TestCodex_ValidateConfig(t *testing.T) {
	cx := &Codex{}
	if err := cx.ValidateConfig(map[string]any{
		"sandbox":          "workspace-write",
		"ask_for_approval": "on-request",
		"full_auto":        true,
	}); err != nil {
		t.Errorf("good config: %v", err)
	}

	err := cx.ValidateConfig(map[string]any{"sandbox": "wide-open"})
	if err == nil || !strings.Contains(err.Error(), "sandbox") {
		t.Errorf("expected sandbox enum error, got %v", err)
	}

	err = cx.ValidateConfig(map[string]any{"ask_for_approval": "maybe"})
	if err == nil || !strings.Contains(err.Error(), "ask_for_approval") {
		t.Errorf("expected ask_for_approval enum error, got %v", err)
	}

	err = cx.ValidateConfig(map[string]any{"full_auto": "yes"})
	if err == nil || !strings.Contains(err.Error(), "must be a bool") {
		t.Errorf("expected bool type error, got %v", err)
	}
}
