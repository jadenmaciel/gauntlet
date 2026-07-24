// Package thresholds implements a monotonic ratchet for quality metrics
// backed by a YAML file. Each metric has a direction (min floor or max
// ceiling) and a value that may only move in the safe direction: floors
// only rise, ceilings only fall.
package thresholds

import (
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"
)

// Direction constants for a metric's ratchet behavior.
const (
	DirectionMin = "min"
	DirectionMax = "max"
)

// Metric is a single named threshold: a floor (min) or ceiling (max).
type Metric struct {
	Direction string  `yaml:"direction"`
	Value     float64 `yaml:"value"`
}

// File is the on-disk thresholds document.
type File struct {
	Metrics map[string]Metric `yaml:"metrics"`
}

// Load reads and parses a thresholds file from path.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading thresholds file %q: %w", path, err)
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing thresholds file %q: %w", path, err)
	}
	return &f, nil
}

// Save writes the thresholds file to path.
func Save(path string, f *File) error {
	data, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("encoding thresholds file: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing thresholds file %q: %w", path, err)
	}
	return nil
}

// Verify checks measured against the named metric's threshold without
// modifying it. It returns an error if the metric is unknown or the
// measured value violates its direction.
func Verify(f *File, metric string, measured float64) error {
	m, ok := f.Metrics[metric]
	if !ok {
		return fmt.Errorf("unknown metric %q", metric)
	}
	if err := verifyValues(metric, m, measured); err != nil {
		return err
	}
	switch m.Direction {
	case DirectionMin:
		if measured < m.Value {
			return fmt.Errorf("%s %v below floor %v", metric, measured, m.Value)
		}
	case DirectionMax:
		if measured > m.Value {
			return fmt.Errorf("%s %v exceeds ceiling %v", metric, measured, m.Value)
		}
	default:
		return fmt.Errorf("metric %q has unknown direction %q", metric, m.Direction)
	}
	return nil
}

func verifyValues(metric string, m Metric, measured float64) error {
	if err := validateMetric(metric, m); err != nil {
		return err
	}
	if !isFinite(measured) {
		return fmt.Errorf("%s has invalid measured value %v", metric, measured)
	}
	return nil
}

// Ratchet moves the named metric's threshold toward measured if that
// move is safe (floor rises, ceiling falls), and refuses otherwise.
// On refusal the threshold is left unchanged and an error is returned.
func Ratchet(f *File, metric string, measured float64) (newValue float64, changed bool, err error) {
	m, ok := f.Metrics[metric]
	if !ok {
		return 0, false, fmt.Errorf("unknown metric %q", metric)
	}
	if err := validateMetric(metric, m); err != nil {
		return 0, false, err
	}
	if !isFinite(measured) {
		return m.Value, false, fmt.Errorf("%s has invalid measured value %v", metric, measured)
	}
	newValue, changed, err = ratchetedValue(metric, m, measured)
	if err != nil || !changed {
		return newValue, changed, err
	}
	m.Value = newValue
	f.Metrics[metric] = m
	return newValue, true, nil
}

func ratchetedValue(name string, metric Metric, measured float64) (float64, bool, error) {
	if metric.Direction == DirectionMin && measured < metric.Value {
		return metric.Value, false, fmt.Errorf("%s: would lower floor from %v to %v", name, metric.Value, measured)
	}
	if metric.Direction == DirectionMax && measured > metric.Value {
		return metric.Value, false, fmt.Errorf("%s: would raise ceiling from %v to %v", name, metric.Value, measured)
	}
	if measured == metric.Value {
		return metric.Value, false, nil
	}
	return measured, true, nil
}

func validateMetric(name string, metric Metric) error {
	if metric.Direction != DirectionMin && metric.Direction != DirectionMax {
		return fmt.Errorf("metric %q has unknown direction %q", name, metric.Direction)
	}
	if !isFinite(metric.Value) {
		return fmt.Errorf("metric %q has invalid value %v", name, metric.Value)
	}
	return nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
