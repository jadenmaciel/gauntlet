// Package crap computes CRAP (Change Risk Anti-Patterns) scores for Go
// functions by combining cyclomatic complexity with test coverage.
package crap

// FunctionComplexity is the cyclomatic complexity of one function or
// method declaration.
type FunctionComplexity struct {
	File       string
	Line       int
	Func       string
	Complexity int
}

// FunctionReport is one function's complexity, coverage, and computed
// CRAP score joined together.
type FunctionReport struct {
	File       string
	Line       int
	Func       string
	Complexity int
	Coverage   float64
	CRAP       float64
	Pass       bool
	// Matched is false when no coverage entry existed for this function
	// and Coverage was therefore defaulted to 0.
	Matched bool
}

// Summary aggregates a Report's function results.
type Summary struct {
	Total   int
	Failing int
	MaxCRAP float64
}

// Report is the result of evaluating a set of functions against coverage
// data and a CRAP ceiling.
type Report struct {
	Ceiling   float64
	Functions []FunctionReport
	Summary   Summary
}
