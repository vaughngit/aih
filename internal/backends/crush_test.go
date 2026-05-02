package backends

import (
	"strings"
	"testing"

	"github.com/vaughngit/aih/internal/manifest"
)

func TestCrush_Registered(t *testing.T) {
	b, err := Get("crush")
	if err != nil {
		t.Fatal(err)
	}
	if !b.SupportsResume() {
		t.Errorf("crush should support resume")
	}
	if b.SupportsAddDir() {
		t.Errorf("crush should not support add-dir")
	}
}

func TestCrush_BuildCommand(t *testing.T) {
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
			[]string{"crush"},
		},
		{
			"resume",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{Resume: true},
			[]string{"crush", "--continue"},
		},
		{
			"all_config_keys",
			&manifest.Manifest{
				Name: "x",
				BackendCfg: map[string]any{
					"crush": map[string]any{
						"binary":   "/opt/homebrew/bin/crush",
						"data_dir": "/tmp/crush",
						"yolo":     true,
						"debug":    true,
						"host":     "unix:///tmp/crush.sock",
					},
				},
			},
			LaunchOpts{},
			[]string{"/opt/homebrew/bin/crush",
				"--data-dir", "/tmp/crush",
				"--yolo",
				"--debug",
				"--host", "unix:///tmp/crush.sock",
			},
		},
		{
			"resources_ignored_silently",
			&manifest.Manifest{Name: "x", Resources: []string{"/a", "/b"}},
			LaunchOpts{},
			[]string{"crush"},
		},
		{
			"passthru_for_specific_session",
			&manifest.Manifest{Name: "x"},
			LaunchOpts{PassThru: []string{"-s", "abc123"}},
			[]string{"crush", "-s", "abc123"},
		},
	}

	c := &Crush{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.BuildCommand(tc.m, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !equalSlice(got, tc.want) {
				t.Errorf("got %v\nwant %v", got, tc.want)
			}
		})
	}
}

func TestCrush_ValidateConfig(t *testing.T) {
	c := &Crush{}
	if err := c.ValidateConfig(map[string]any{"yolo": true, "data_dir": "/tmp"}); err != nil {
		t.Errorf("good config: %v", err)
	}
	err := c.ValidateConfig(map[string]any{"yolo": "yes"})
	if err == nil || !strings.Contains(err.Error(), "must be a bool") {
		t.Errorf("expected bool type error, got %v", err)
	}
}
