package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

type cliResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

type fakeRecord struct {
	Argv   []string          `json:"argv"`
	Env    map[string]string `json:"env"`
	Cwd    string            `json:"cwd"`
	Signal string            `json:"signal,omitempty"`
}

type exitPanic int

var (
	fakeBackendOnce sync.Once
	fakeBackendPath string
	fakeBackendErr  error
)

func setupAihHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AIH_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeManifest(t *testing.T, name, body string) {
	t.Helper()
	dir, err := registryDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".toml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runAih(t *testing.T, args ...string) cliResult {
	t.Helper()
	var stdout, stderr bytes.Buffer
	oldExit := exit
	exit = func(code int) {
		panic(exitPanic(code))
	}
	defer func() { exit = oldExit }()

	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	res := cliResult{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				code, ok := r.(exitPanic)
				if !ok {
					panic(r)
				}
				res.ExitCode = int(code)
			}
		}()
		if err := cmd.Execute(); err != nil {
			res.Err = err
			res.ExitCode = 1
			fmt.Fprintln(&stderr, "aih: "+err.Error())
		}
	}()

	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	return res
}

func buildFakeBackend(t *testing.T) string {
	t.Helper()
	fakeBackendOnce.Do(func() {
		dir, err := os.MkdirTemp("", "aih-fakebackend-*")
		if err != nil {
			fakeBackendErr = err
			return
		}
		fakeBackendPath = filepath.Join(dir, "fakebackend")
		if runtime.GOOS == "windows" {
			fakeBackendPath += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", fakeBackendPath, "./internal/testutil/fakebackend")
		cmd.Dir = repoRoot(t)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fakeBackendErr = fmt.Errorf("go build fakebackend: %w\n%s", err, out)
		}
	})
	if fakeBackendErr != nil {
		t.Fatal(fakeBackendErr)
	}
	return fakeBackendPath
}

func recordedRun(t *testing.T, path string) fakeRecord {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rec fakeRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	return rec
}

func registryDir() (string, error) {
	home := os.Getenv("AIH_HOME")
	if home == "" {
		return "", fmt.Errorf("AIH_HOME is not set")
	}
	return filepath.Join(home, "agents"), nil
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
