package thresholds

import (
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerify(t *testing.T) {
	tf := &File{
		Metrics: map[string]Metric{
			"mutation_efficacy": {Direction: "min", Value: 50},
			"crap_ceiling":      {Direction: "max", Value: 8},
		},
	}

	cases := []struct {
		name     string
		metric   string
		measured float64
		wantErr  string // "" means no error
	}{
		{"min pass above floor", "mutation_efficacy", 60, ""},
		{"min pass exact boundary", "mutation_efficacy", 50, ""},
		{"min fail below floor", "mutation_efficacy", 49, "mutation_efficacy 49 below floor 50"},
		{"max pass below ceiling", "crap_ceiling", 5, ""},
		{"max pass exact boundary", "crap_ceiling", 8, ""},
		{"max fail above ceiling", "crap_ceiling", 9, "crap_ceiling 9 exceeds ceiling 8"},
		{"unknown metric", "does_not_exist", 1, "unknown metric"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Verify(tf, c.metric, c.measured)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("expected error containing %q, got %q", c.wantErr, err.Error())
			}
		})
	}
}

func TestRatchet(t *testing.T) {
	cases := []struct {
		name        string
		direction   string
		start       float64
		measured    float64
		wantValue   float64
		wantChanged bool
		wantErr     string
	}{
		{"min raises floor", "min", 50, 60, 60, true, ""},
		{"min no-op when equal", "min", 50, 50, 50, false, ""},
		{"min refuses to lower floor", "min", 50, 40, 50, false, "would lower floor"},
		{"max lowers ceiling", "max", 8, 5, 5, true, ""},
		{"max no-op when equal", "max", 8, 8, 8, false, ""},
		{"max refuses to raise ceiling", "max", 8, 9, 8, false, "would raise ceiling"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tf := &File{
				Metrics: map[string]Metric{
					"m": {Direction: c.direction, Value: c.start},
				},
			}
			newVal, changed, err := Ratchet(tf, "m", c.measured)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", c.wantErr)
				}
				if !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("expected error containing %q, got %q", c.wantErr, err.Error())
				}
			}
			if changed != c.wantChanged {
				t.Fatalf("expected changed=%v, got %v", c.wantChanged, changed)
			}
			if newVal != c.wantValue {
				t.Fatalf("expected newValue=%v, got %v", c.wantValue, newVal)
			}
			if tf.Metrics["m"].Value != c.wantValue {
				t.Fatalf("expected stored value=%v, got %v", c.wantValue, tf.Metrics["m"].Value)
			}
		})
	}
}

func TestRatchetUnknownMetric(t *testing.T) {
	tf := &File{Metrics: map[string]Metric{}}
	_, changed, err := Ratchet(tf, "nope", 1)
	if err == nil {
		t.Fatal("expected error for unknown metric")
	}
	if changed {
		t.Fatal("expected changed=false for unknown metric")
	}
}

func TestRatchetRejectsNonFiniteMeasuredValue(t *testing.T) {
	tf := &File{Metrics: map[string]Metric{
		"coverage_min": {Direction: DirectionMin, Value: 80},
	}}

	_, changed, err := Ratchet(tf, "coverage_min", math.NaN())
	if err == nil {
		t.Fatal("Ratchet() error = nil, want error for NaN")
	}
	if changed {
		t.Fatal("Ratchet() changed = true, want false for NaN")
	}
	if got := tf.Metrics["coverage_min"].Value; got != 80 {
		t.Fatalf("metric value = %v, want 80", got)
	}
}

func TestVerifyRejectsInvalidMetricAndMeasuredValue(t *testing.T) {
	invalid := &File{Metrics: map[string]Metric{"bad": {Direction: "bad", Value: 1}}}
	if err := Verify(invalid, "bad", 1); err == nil {
		t.Fatal("Verify() error = nil, want invalid direction error")
	}
	valid := &File{Metrics: map[string]Metric{"coverage_min": {Direction: DirectionMin, Value: 80}}}
	if err := Verify(valid, "coverage_min", math.NaN()); err == nil {
		t.Fatal("Verify() error = nil, want NaN error")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "thresholds.yml")

	orig := &File{
		Metrics: map[string]Metric{
			"crap_ceiling":      {Direction: "max", Value: 8},
			"mutation_efficacy": {Direction: "min", Value: 0},
		},
	}

	if err := Save(path, orig); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Metrics) != len(orig.Metrics) {
		t.Fatalf("expected %d metrics, got %d", len(orig.Metrics), len(loaded.Metrics))
	}
	for name, m := range orig.Metrics {
		got, ok := loaded.Metrics[name]
		if !ok {
			t.Fatalf("missing metric %q after round trip", name)
		}
		if got.Direction != m.Direction || got.Value != m.Value {
			t.Fatalf("metric %q mismatch: got %+v, want %+v", name, got, m)
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yml"))
	if err == nil {
		t.Fatal("expected error loading missing file")
	}
}
