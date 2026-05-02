package backends

import (
	"strings"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestKiro_Registered(t *testing.T) {
	b, err := Get("kiro")
	if err != nil {
		t.Fatal(err)
	}
	if b.SupportsResume() {
		t.Errorf("kiro should not support resume")
	}
	if !b.SupportsAddDir() {
		t.Errorf("kiro should support add (mapped to --add-file)")
	}
}

func TestKiro_BuildCommand(t *testing.T) {
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
			[]string{"kiro", "chat"},
		},
		{
			"all_config_keys",
			&manifest.Manifest{
				Name: "x",
				BackendCfg: map[string]any{
					"kiro": map[string]any{
						"binary":       "/usr/local/bin/kiro",
						"mode":         "agent",
						"profile":      "work",
						"new_window":   true,
						"reuse_window": false,
						"maximize":     true,
					},
				},
			},
			LaunchOpts{},
			[]string{"/usr/local/bin/kiro", "chat",
				"--mode", "agent",
				"--profile", "work",
				"--new-window",
				"--maximize",
			},
		},
		{
			"resources_become_add_file",
			&manifest.Manifest{Name: "x", Resources: []string{"/a.md", "/b.md"}},
			LaunchOpts{AddDirs: []string{"/c.md"}},
			[]string{"kiro", "chat",
				"--add-file", "/a.md",
				"--add-file", "/b.md",
				"--add-file", "/c.md",
			},
		},
		{
			"passthru_after_flags",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{PassThru: []string{"explain this"}},
			[]string{"kiro", "chat", "explain this"},
		},
	}

	k := &Kiro{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := k.BuildCommand(tc.m, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !equalSlice(got, tc.want) {
				t.Errorf("got %v\nwant %v", got, tc.want)
			}
		})
	}
}

func TestKiro_ValidateConfig(t *testing.T) {
	k := &Kiro{}
	if err := k.ValidateConfig(map[string]any{
		"mode":     "agent",
		"maximize": true,
	}); err != nil {
		t.Errorf("good config: %v", err)
	}
	err := k.ValidateConfig(map[string]any{"unknown_key": "x"})
	if err == nil || !strings.Contains(err.Error(), "unknown_key") {
		t.Errorf("expected unknown key error, got %v", err)
	}
}
