package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSyncDataPassesArgumentsToScript(t *testing.T) {
	dir := t.TempDir()
	recorder := filepath.Join(dir, "python")
	output := filepath.Join(dir, "args.txt")
	script := filepath.Join(dir, "sync.py")
	if err := os.WriteFile(script, []byte("# test placeholder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	recorderBody := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + output + "\n"
	if err := os.WriteFile(recorder, []byte(recorderBody), 0o755); err != nil {
		t.Fatal(err)
	}

	err := run([]string{
		"rst",
		"sync-data",
		"-python", recorder,
		"-script", script,
		"-provider", "auto",
		"-universe", "data/universe/ai_tech.csv",
		"-out", "data/live",
		"-start", "20240101",
		"-end", "20240601",
		"-adjust", "qfq",
	})
	if err != nil {
		t.Fatalf("sync-data returned error: %v", err)
	}

	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	for _, want := range []string{
		script,
		"--provider\nauto",
		"--universe\ndata/universe/ai_tech.csv",
		"--out\ndata/live",
		"--start\n20240101",
		"--end\n20240601",
		"--adjust\nqfq",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected recorded args to contain %q, got:\n%s", want, got)
		}
	}
}
