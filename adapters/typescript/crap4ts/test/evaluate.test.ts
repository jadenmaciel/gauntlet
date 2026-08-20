import assert from "node:assert/strict";
import test from "node:test";

import { evaluate } from "../src/evaluate.js";

test("evaluate keeps functions missing coverage with coverage 0 and matched false", () => {
  const report = evaluate(
    [
      {
        file: "src/example.ts",
        line: 12,
        func: "hotPath",
        complexity: 10,
      },
    ],
    new Map<string, number>(),
    30,
  );

  assert.equal(report.functions.length, 1);
  assert.equal(report.summary.total, 1);
  assert.equal(report.functions[0].coverage, 0);
  assert.equal(report.functions[0].matched, false);
  assert.equal(report.functions[0].pass, false);
});

test("evaluate reports matched coverage as a percentage", () => {
  const report = evaluate(
    [
      {
        file: "src/example.ts",
        line: 12,
        func: "hotPath",
        complexity: 3,
      },
    ],
    new Map([["src/example.ts:12:hotPath", 0.18]]),
    8,
  );

  assert.equal(report.functions[0].coverage, 18);
  assert.ok(Math.abs(report.functions[0].crap - 7.962312) < 1e-9);
  assert.equal(report.functions[0].matched, true);
});
