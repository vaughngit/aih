package backends

import (
	"strings"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestOpenCode_Registered(t *testing.T) {
	b, err := Get("opencode")
	if err != nil {
		t.Fatal(err)
	}
	if !b.SupportsResume() {
		t.Errorf("opencode should support resume")
	}
	if b.SupportsAddDir() {
		t.Errorf("opencode should not support add-dir")
	}
}

func TestOpenCode_BuildCommand(t *testing.T) {
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
			[]string{"opencode"},
		},
		{
			"resume",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{Resume: true},
			[]string{"opencode", "--continue"},
		},
		{
			"all_config_keys",
			&manifest.Manifest{
				Name: "x",
				BackendCfg: map[string]any{
					"opencode": map[string]any{
						"binary":    "/Users/me/.opencode/bin/opencode",
						"model":     "anthropic/claude-opus-4-7",
						"agent":     "build",
						"log_level": "WARN",
						"pure":      true,
					},
				},
			},
			LaunchOpts{},
			[]string{"/Users/me/.opencode/bin/opencode",
				"--model", "anthropic/claude-opus-4-7",
				"--agent", "build",
				"--log-level", "WARN",
				"--pure",
			},
		},
		{
			"resources_ignored_silently",
			&manifest.Manifest{Name: "x", Resources: []string{"/a"}},
			LaunchOpts{},
			[]string{"opencode"},
		},
		{
			"passthru_for_specific_session",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{PassThru: []string{"--session", "abc123"}},
			[]string{"opencode", "--session", "abc123"},
		},
	}

	oc := &OpenCode{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := oc.BuildCommand(tc.m, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !equalSlice(got, tc.want) {
				t.Errorf("got %v\nwant %v", got, tc.want)
			}
		})
	}
}

func TestOpenCode_ValidateConfig(t *testing.T) {
	oc := &OpenCode{}
	if err := oc.ValidateConfig(map[string]any{"log_level": "INFO"}); err != nil {
		t.Errorf("good config: %v", err)
	}
	err := oc.ValidateConfig(map[string]any{"log_level": "verbose"})
	if err == nil || !strings.Contains(err.Error(), "log_level") {
		t.Errorf("expected log_level enum error, got %v", err)
	}
}
