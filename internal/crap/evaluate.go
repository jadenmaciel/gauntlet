package crap

import "fmt"

// buildKey is the shared "file:line:funcname" key format used to join
// complexity results with coverage data.
func buildKey(file string, line int, name string) string {
	return fmt.Sprintf("%s:%d:%s", file, line, name)
}

// Evaluate joins complexity and coverage data by file:line:funcname,
// computes each function's CRAP score, and summarizes pass/fail counts.
// A function with no matching coverage entry gets Coverage 0 and
// Matched false rather than being dropped.
func Evaluate(functions []FunctionComplexity, coverage map[string]float64, ceiling float64) Report {
	report := Report{Ceiling: ceiling}

	var maxCRAP float64
	failing := 0
	for _, fn := range functions {
		cov, matched := coverage[buildKey(fn.File, fn.Line, fn.Func)]
		score := ComputeCRAP(fn.Complexity, cov)
		pass := score <= ceiling
		if !pass {
			failing++
		}
		if score > maxCRAP {
			maxCRAP = score
		}

		report.Functions = append(report.Functions, FunctionReport{
			File:       fn.File,
			Line:       fn.Line,
			Func:       fn.Func,
			Complexity: fn.Complexity,
			Coverage:   cov,
			CRAP:       score,
			Pass:       pass,
			Matched:    matched,
		})
	}

	report.Summary = Summary{
		Total:   len(functions),
		Failing: failing,
		MaxCRAP: maxCRAP,
	}
	return report
}
