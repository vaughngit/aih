package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGoldens = flag.Bool("update", false, "update golden files")

func TestShowGolden(t *testing.T) {
	cases := []string{"minimal", "full-claude", "expanded-paths"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			setupAihHome(t)
			t.Setenv("HOME", "/tmp/aih-test-home")
			t.Setenv("AIH_SAMPLE_ROOT", "/tmp/aih-sample-root")
			b, err := os.ReadFile(filepath.Join("testdata", "manifests", name+".toml"))
			if err != nil {
				t.Fatal(err)
			}
			writeManifest(t, name, string(b))

			res := runAih(t, "show", name)
			if res.ExitCode != 0 || res.Err != nil {
				t.Fatalf("show failed: exit=%d err=%v stderr=%s", res.ExitCode, res.Err, res.Stderr)
			}
			if !strings.Contains(res.Stderr, "# source: ") {
				t.Fatalf("stderr = %q, want source line", res.Stderr)
			}

			golden := filepath.Join("testdata", "show", name+".golden")
			if *updateGoldens {
				if err := os.WriteFile(golden, []byte(res.Stdout), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if res.Stdout != string(want) {
				t.Fatalf("show output mismatch for %s\n--- got ---\n%s--- want ---\n%s", name, res.Stdout, want)
			}
		})
	}
}
