import assert from "node:assert/strict";
import test from "node:test";

import { computeCRAP } from "../src/crap.js";

const EPSILON = 1e-9;

const vectors: Array<{ complexity: number; coveragePercent: number; expected: number }> = [
  { complexity: 1, coveragePercent: 0, expected: 2 },
  { complexity: 1, coveragePercent: 100, expected: 1 },
  { complexity: 2, coveragePercent: 0, expected: 6 },
  { complexity: 3, coveragePercent: 50, expected: 4.125 },
  { complexity: 5, coveragePercent: 0, expected: 30 },
  { complexity: 5, coveragePercent: 50, expected: 8.125 },
  { complexity: 5, coveragePercent: 100, expected: 5 },
  { complexity: 8, coveragePercent: 100, expected: 8 },
  { complexity: 10, coveragePercent: 0, expected: 110 },
  { complexity: 10, coveragePercent: 80, expected: 10.8 },
  { complexity: 10, coveragePercent: 100, expected: 10 },
  { complexity: 15, coveragePercent: 100, expected: 15 },
  { complexity: 20, coveragePercent: 90, expected: 20.4 },
];

test("computeCRAP matches locked golden vectors", () => {
  for (const vector of vectors) {
    const actual = computeCRAP(vector.complexity, vector.coveragePercent);
    assert.ok(
      Math.abs(actual - vector.expected) < EPSILON,
      `CC=${vector.complexity} coverage=${vector.coveragePercent}% got=${actual} expected=${vector.expected}`,
    );
  }
});
