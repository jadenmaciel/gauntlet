package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/thresholds"
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

func TestRunInitWritesTheSuppliedMeasuredCeiling(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := run([]string{"init", "--crap-ceiling", "42.5", path}); err != nil {
		t.Fatalf("run(init) error = %v", err)
	}

	file, err := thresholds.Load(path)
	if err != nil {
		t.Fatalf("thresholds.Load() error = %v", err)
	}
	metric, ok := file.Metrics["crap_ceiling"]
	if !ok {
		t.Fatal("generated file has no crap_ceiling metric")
	}
	if metric.Value != 42.5 {
		t.Errorf("crap_ceiling = %v, want 42.5", metric.Value)
	}
	if metric.Direction != "max" {
		t.Errorf("direction = %q, want max", metric.Direction)
	}

	body := readFile(t, path)
	if strings.Contains(body, "PLACEHOLDER") {
		t.Error("a measured ceiling must not be labelled a placeholder")
	}
}

func TestRunInitMarksTheDefaultCeilingAsAPlaceholder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thresholds.yml")
	if err := run([]string{"init", path}); err != nil {
		t.Fatalf("run(init) error = %v", err)
	}

	body := readFile(t, path)
	if !strings.Contains(body, "PLACEHOLDER") {
		t.Error("default ceiling is not marked as a placeholder")
	}

	file, err := thresholds.Load(path)
	if err != nil {
		t.Fatalf("thresholds.Load() error = %v", err)
	}
	if got := file.Metrics["crap_ceiling"].Value; got != 8 {
		t.Errorf("crap_ceiling = %v, want 8", got)
	}
}

func TestRunInitRejectsANonPositiveCeiling(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		path := filepath.Join(t.TempDir(), "thresholds.yml")
		if err := run([]string{"init", "--crap-ceiling", value, path}); err == nil {
			t.Errorf("run(init --crap-ceiling %s) error = nil, want rejection", value)
		}
	}
}

func TestFormatCeilingDropsATrailingZero(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  string
	}{{8, "8"}, {42.5, "42.5"}, {110, "110"}} {
		if got := formatCeiling(tc.value); got != tc.want {
			t.Errorf("formatCeiling(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
