package crap_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func TestComputeCRAP_GoldenVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/crap-golden.json")
	if err != nil {
		t.Fatal(err)
	}

	var vectors []struct {
		Name       string  `json:"name"`
		Complexity int     `json:"complexity"`
		Coverage   float64 `json:"coverage"`
		CRAP       float64 `json:"crap"`
	}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}

	for _, vector := range vectors {
		t.Run(vector.Name, func(t *testing.T) {
			got := crap.ComputeCRAP(vector.Complexity, vector.Coverage)
			if math.Abs(got-vector.CRAP) > 1e-9 {
				t.Errorf("ComputeCRAP(%d, %v) = %v, want %v", vector.Complexity, vector.Coverage, got, vector.CRAP)
			}
		})
	}
}
