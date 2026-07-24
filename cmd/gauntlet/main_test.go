package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunVerifyReturnsViolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := os.WriteFile(path, []byte("metrics:\n  coverage_min:\n    direction: min\n    value: 80\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := runVerify([]string{"--thresholds", path, "--metric", "coverage_min", "--value", "79"})
	if err == nil {
		t.Fatal("runVerify() error = nil, want threshold violation")
	}
}

func TestRunInitAndRatchet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := run([]string{"init", path}); err != nil {
		t.Fatalf("run(init) error = %v", err)
	}
	if err := run([]string{"ratchet", "--thresholds", path, "--metric", "crap_ceiling", "--value", "7"}); err != nil {
		t.Fatalf("run(ratchet) error = %v", err)
	}
	if err := run([]string{"verify", "--thresholds", path, "--metric", "crap_ceiling", "--value", "7"}); err != nil {
		t.Fatalf("run(verify) error = %v", err)
	}
}

func TestRunRejectsUnknownSubcommand(t *testing.T) {
	if err := run([]string{"unknown"}); err == nil {
		t.Fatal("run() error = nil, want unknown subcommand error")
	}
}

func TestRunRatchetRejectsMissingMetric(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := os.WriteFile(path, []byte("metrics:\n  coverage_min:\n    direction: min\n    value: 80\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"ratchet", "--thresholds", path, "--metric", "missing", "--value", "80"}); err == nil {
		t.Fatal("run(ratchet) error = nil, want unknown metric error")
	}
}

func TestRunInitRefusesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"init", path}); err == nil {
		t.Fatal("run(init) error = nil, want existing file error")
	}
}

func TestRunInitForceOverwritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"init", "--force", path}); err != nil {
		t.Fatalf("run(init) error = %v", err)
	}
}

func TestRunRatchetRejectsExtraArgument(t *testing.T) {
	if err := run([]string{"ratchet", "extra"}); err == nil {
		t.Fatal("run(ratchet) error = nil, want flags-only error")
	}
}
