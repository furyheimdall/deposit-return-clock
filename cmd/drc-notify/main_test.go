package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunListAndFire(t *testing.T) {
	dir := t.TempDir()
	cases := filepath.Join(dir, "cases.json")
	body := `[{"id":"ca-1","state":"CA","vacated_on":"2026-06-01","lease_ends_on":"2026-06-01"}]`
	if err := os.WriteFile(cases, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "fired.jsonl")

	stdout := filepath.Join(dir, "stdout.txt")
	f, err := os.Create(stdout)
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = f
	err = run([]string{
		"-cases", cases,
		"-today", "2026-06-15",
		"-within", "7",
		"-fire",
		"-notifier", "file",
		"-out", out,
	})
	os.Stdout = old
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	list, err := os.ReadFile(stdout)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(list), "ca-1\tCA\tdeadline=2026-06-22\tdays=7") {
		t.Fatalf("list: %s", list)
	}
	if !strings.Contains(string(list), "fired=1") {
		t.Fatalf("fired count: %s", list)
	}
	fired, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fired), `"kind":"D-7"`) {
		t.Fatalf("file notifier: %s", fired)
	}
}

func TestRunRequiresCases(t *testing.T) {
	if err := run(nil); err == nil {
		t.Fatal("expected error")
	}
}
