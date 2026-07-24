package crap

// ComputeCRAP implements CRAP = CC^2 * (1-coverage)^3 + CC.
//
// coveragePercent is 0-100 and is converted to a 0-1 fraction internally.
func ComputeCRAP(complexity int, coveragePercent float64) float64 {
	cc := float64(complexity)
	coverage := coveragePercent / 100
	uncovered := 1 - coverage
	return cc*cc*uncovered*uncovered*uncovered + cc
}
