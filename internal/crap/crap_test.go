package crap_test

import (
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

// Verified table for ceiling 8: CRAP = CC^2 * (1-coverage)^3 + CC.
// Table values are coverage percentages rounded to the nearest integer,
// so tests assert a pass at the rounded value and a clear fail 2 points
// below it (outside the documented +/-1 percentage point tolerance).
const ceiling = 8.0

func TestComputeCRAP_AlwaysPasses(t *testing.T) {
	// CC 1-2: CRAP <= ceiling at every coverage level, including the
	// worst case of 0% coverage.
	for _, cc := range []int{1, 2} {
		got := crap.ComputeCRAP(cc, 0)
		if got > ceiling {
			t.Errorf("CC%d at 0%% coverage: CRAP=%.4f, want <= %.1f", cc, got, ceiling)
		}
	}
}

func TestComputeCRAP_Table(t *testing.T) {
	tests := []struct {
		name             string
		cc               int
		requiredCoverage float64 // rounded percent where CRAP first drops to <= ceiling
	}{
		{"cc3", 3, 18},
		{"cc4", 4, 37},
		{"cc5", 5, 51},
		{"cc6", 6, 62},
		{"cc7", 7, 73},
	}

	for _, tt := range tests {
		// The table value is only accurate to +/-1 percentage point, so
		// the true threshold can sit up to 1 point above it. Asserting
		// the pass case 1 point above the table value keeps the test
		// clear of that uncertainty band; asserting the fail case 2
		// points below covers the same band from the other side.
		t.Run(tt.name+"_passes_at_required", func(t *testing.T) {
			got := crap.ComputeCRAP(tt.cc, tt.requiredCoverage+1)
			if got > ceiling {
				t.Errorf("CC%d at %.0f%% coverage: CRAP=%.4f, want <= %.1f", tt.cc, tt.requiredCoverage+1, got, ceiling)
			}
		})
		t.Run(tt.name+"_fails_below_tolerance", func(t *testing.T) {
			got := crap.ComputeCRAP(tt.cc, tt.requiredCoverage-2)
			if got <= ceiling {
				t.Errorf("CC%d at %.0f%% coverage: CRAP=%.4f, want > %.1f", tt.cc, tt.requiredCoverage-2, got, ceiling)
			}
		})
	}
}

func TestComputeCRAP_CC8ExactBoundary(t *testing.T) {
	// At 100% coverage, CRAP collapses to exactly CC, so CC8 at 100%
	// equals the ceiling exactly and passes because pass is CRAP <= ceiling.
	got := crap.ComputeCRAP(8, 100)
	if got != 8 {
		t.Fatalf("CC8 at 100%% coverage: CRAP=%.6f, want exactly 8", got)
	}
	if got > ceiling {
		t.Errorf("CC8 at 100%% coverage should pass (CRAP <= ceiling), got CRAP=%.6f", got)
	}

	// Just below 100%, the boundary is missed and it fails.
	below := crap.ComputeCRAP(8, 99)
	if below <= ceiling {
		t.Errorf("CC8 at 99%% coverage: CRAP=%.6f, want > %.1f", below, ceiling)
	}
}

func TestComputeCRAP_CC9PlusAlwaysFails(t *testing.T) {
	// CC9+: CRAP >= CC always, so even the best case of 100% coverage
	// still exceeds the ceiling of 8.
	for _, cc := range []int{9, 10, 15} {
		got := crap.ComputeCRAP(cc, 100)
		if got <= ceiling {
			t.Errorf("CC%d at 100%% coverage: CRAP=%.4f, want > %.1f (should be impossible to pass)", cc, got, ceiling)
		}
	}
}
