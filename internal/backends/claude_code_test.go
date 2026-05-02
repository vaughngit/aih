package backends

import (
	"strings"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestClaudeCode_Registered(t *testing.T) {
	b, err := Get("claude-code")
	if err != nil {
		t.Fatal(err)
	}
	if !b.SupportsResume() || !b.SupportsAddDir() {
		t.Errorf("claude-code should support resume and add-dir")
	}
}

func TestClaudeCode_BuildCommand(t *testing.T) {
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
			[]string{"claude"},
		},
		{
			"resume_maps_to_continue",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{Resume: true},
			[]string{"claude", "--continue"},
		},
		{
			"resources_become_add_dir",
			&manifest.Manifest{Name: "x", Resources: []string{"/a", "/b"}},
			LaunchOpts{},
			[]string{"claude", "--add-dir", "/a", "--add-dir", "/b"},
		},
		{
			"opts_add_dirs_after_resources",
			&manifest.Manifest{Name: "x", Resources: []string{"/a"}},
			LaunchOpts{AddDirs: []string{"/b"}},
			[]string{"claude", "--add-dir", "/a", "--add-dir", "/b"},
		},
		{
			"all_config_keys",
			&manifest.Manifest{
				Name: "x",
				BackendCfg: map[string]any{
					"claude-code": map[string]any{
						"binary":               "/usr/local/bin/claude",
						"agent":                "kubernetes",
						"model":                "claude-opus-4-7",
						"permission_mode":      "acceptEdits",
						"append_system_prompt": "be terse",
					},
				},
			},
			LaunchOpts{},
			[]string{
				"/usr/local/bin/claude",
				"--agent", "kubernetes",
				"--model", "claude-opus-4-7",
				"--permission-mode", "acceptEdits",
				"--append-system-prompt", "be terse",
			},
		},
		{
			"passthru_appended_last",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{Resume: true, PassThru: []string{"--debug", "api"}},
			[]string{"claude", "--continue", "--debug", "api"},
		},
	}

	cc := &ClaudeCode{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cc.BuildCommand(tc.m, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !equalSlice(got, tc.want) {
				t.Errorf("got %v\nwant %v", got, tc.want)
			}
		})
	}
}

func TestClaudeCode_ValidateConfig(t *testing.T) {
	cc := &ClaudeCode{}
	if err := cc.ValidateConfig(nil); err != nil {
		t.Errorf("nil should be ok, got %v", err)
	}
	if err := cc.ValidateConfig(map[string]any{"agent": "x", "model": "y"}); err != nil {
		t.Errorf("known keys should pass, got %v", err)
	}

	err := cc.ValidateConfig(map[string]any{"unknown": "x"})
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected unknown key error, got %v", err)
	}

	err = cc.ValidateConfig(map[string]any{"agent": 42})
	if err == nil || !strings.Contains(err.Error(), "must be a string") {
		t.Errorf("expected type error, got %v", err)
	}

	err = cc.ValidateConfig(map[string]any{"permission_mode": "wide-open"})
	if err == nil || !strings.Contains(err.Error(), "permission_mode") {
		t.Errorf("expected enum error, got %v", err)
	}
}
