package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type record struct {
	Argv   []string          `json:"argv"`
	Env    map[string]string `json:"env"`
	Cwd    string            `json:"cwd"`
	Signal string            `json:"signal,omitempty"`
}

func main() {
	if s := os.Getenv("AIH_FAKE_STDOUT"); s != "" {
		_, _ = fmt.Fprint(os.Stdout, s)
	}
	if s := os.Getenv("AIH_FAKE_STDERR"); s != "" {
		_, _ = fmt.Fprint(os.Stderr, s)
	}

	rec := newRecord()
	if os.Getenv("AIH_FAKE_HANG") == "1" {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		sig := <-ch
		rec.Signal = sig.String()
		writeRecord(rec)
		os.Exit(0)
	}

	writeRecord(rec)
	os.Exit(fakeExit())
}

func newRecord() record {
	cwd, _ := os.Getwd()
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				env[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return record{Argv: os.Args, Env: env, Cwd: cwd}
}

func writeRecord(rec record) {
	path := os.Getenv("AIH_FAKE_RECORD")
	if path == "" {
		return
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "fakebackend: marshal: %v\n", err)
		os.Exit(2)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "fakebackend: write: %v\n", err)
		os.Exit(2)
	}
}

func fakeExit() int {
	raw := os.Getenv("AIH_FAKE_EXIT")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fakebackend: invalid AIH_FAKE_EXIT %q\n", raw)
		return 2
	}
	return n
}
