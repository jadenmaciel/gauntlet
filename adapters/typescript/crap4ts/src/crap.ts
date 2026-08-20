// ComputeCRAP implements CRAP = CC^2 * (1-coverage)^3 + CC.
// coveragePercent is 0-100 and is converted to a 0-1 fraction internally.
export function computeCRAP(complexity: number, coveragePercent: number): number {
  const cc = Number(complexity);
  const coverage = coveragePercent / 100;
  const uncovered = 1 - coverage;
  return cc * cc * uncovered * uncovered * uncovered + cc;
}
