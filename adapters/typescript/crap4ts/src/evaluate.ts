import { computeCRAP } from "./crap.js";
import { buildKey } from "./key.js";
import type { FunctionComplexity, Report } from "./types.js";

export function evaluate(
  functions: FunctionComplexity[],
  coverage: Map<string, number>,
  ceiling: number,
): Report {
  const report: Report = {
    ceiling,
    functions: [],
    summary: {
      total: functions.length,
      failing: 0,
      max_crap: 0,
    },
  };

  for (const fn of functions) {
    const key = buildKey(fn.file, fn.line, fn.func);
    const matched = coverage.has(key);
    const coverageFraction = matched ? clampCoverage(coverage.get(key) ?? 0) : 0;
    const coveragePercent = coverageFraction * 100;
    const score = computeCRAP(fn.complexity, coveragePercent);
    const pass = score <= ceiling;

    if (!pass) {
      report.summary.failing += 1;
    }
    if (score > report.summary.max_crap) {
      report.summary.max_crap = score;
    }

    report.functions.push({
      file: fn.file,
      line: fn.line,
      func: fn.func,
      complexity: fn.complexity,
      coverage: coveragePercent,
      crap: score,
      pass,
      matched,
    });
  }

  return report;
}

function clampCoverage(value: number): number {
  if (Number.isNaN(value)) {
    return 0;
  }
  if (value < 0) {
    return 0;
  }
  if (value > 1) {
    return 1;
  }
  return value;
}
