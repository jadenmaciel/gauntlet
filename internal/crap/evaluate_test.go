package crap_test

import (
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func TestEvaluate_JoinsByFileLineFunc(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "a.go", Line: 10, Func: "Foo", Complexity: 3},
	}
	coverage := map[string]float64{
		"a.go:10:Foo": 18,
	}

	report := crap.Evaluate(functions, coverage, 8)

	if len(report.Functions) != 1 {
		t.Fatalf("got %d function reports, want 1", len(report.Functions))
	}
	fr := report.Functions[0]
	if fr.Coverage != 18 {
		t.Errorf("Coverage = %v, want 18", fr.Coverage)
	}
	if !fr.Matched {
		t.Errorf("Matched = false, want true (coverage entry present)")
	}
	wantCRAP := crap.ComputeCRAP(3, 18)
	if fr.CRAP != wantCRAP {
		t.Errorf("CRAP = %v, want %v", fr.CRAP, wantCRAP)
	}
	if !fr.Pass {
		t.Errorf("Pass = false, want true at CC3/18%%")
	}
}

func TestEvaluate_UnmatchedCoverageDefaultsToZero(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "a.go", Line: 10, Func: "Foo", Complexity: 2},
	}
	// No coverage data at all for this function.
	report := crap.Evaluate(functions, map[string]float64{}, 8)

	fr := report.Functions[0]
	if fr.Coverage != 0 {
		t.Errorf("Coverage = %v, want 0 for unmatched function", fr.Coverage)
	}
	if fr.Matched {
		t.Errorf("Matched = true, want false for unmatched function")
	}
}

func TestEvaluate_PassFailBoundary(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "a.go", Line: 1, Func: "Passing", Complexity: 8},
		{File: "a.go", Line: 2, Func: "Failing", Complexity: 8},
	}
	coverage := map[string]float64{
		"a.go:1:Passing": 100, // CRAP exactly 8, passes (ceiling is <=)
		"a.go:2:Failing": 99,  // CRAP just above 8, fails
	}

	report := crap.Evaluate(functions, coverage, 8)

	byName := map[string]crap.FunctionReport{}
	for _, fr := range report.Functions {
		byName[fr.Func] = fr
	}

	if !byName["Passing"].Pass {
		t.Errorf("Passing: Pass = false, want true")
	}
	if byName["Failing"].Pass {
		t.Errorf("Failing: Pass = true, want false")
	}
}

func TestEvaluate_Summary(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "a.go", Line: 1, Func: "Ok", Complexity: 1},
		{File: "a.go", Line: 2, Func: "BadOne", Complexity: 9},
		{File: "a.go", Line: 3, Func: "BadTwo", Complexity: 10},
	}
	coverage := map[string]float64{
		"a.go:1:Ok":     0,
		"a.go:2:BadOne": 100,
		"a.go:3:BadTwo": 100,
	}

	report := crap.Evaluate(functions, coverage, 8)

	if report.Summary.Total != 3 {
		t.Errorf("Summary.Total = %d, want 3", report.Summary.Total)
	}
	if report.Summary.Failing != 2 {
		t.Errorf("Summary.Failing = %d, want 2", report.Summary.Failing)
	}
	wantMax := crap.ComputeCRAP(10, 100)
	if report.Summary.MaxCRAP != wantMax {
		t.Errorf("Summary.MaxCRAP = %v, want %v", report.Summary.MaxCRAP, wantMax)
	}
	if report.Ceiling != 8 {
		t.Errorf("Report.Ceiling = %v, want 8", report.Ceiling)
	}
}

func TestEvaluate_NoFunctions(t *testing.T) {
	report := crap.Evaluate(nil, map[string]float64{}, 8)
	if report.Summary.Total != 0 || report.Summary.Failing != 0 || report.Summary.MaxCRAP != 0 {
		t.Errorf("empty input: got %+v, want all zero summary", report.Summary)
	}
}
