import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { readCoverage } from "../src/coverage.js";
import type { FunctionComplexity } from "../src/types.js";

const method: FunctionComplexity = {
  file: "src/example.ts",
  line: 2,
  func: "Example.method",
  complexity: 2,
};

test("Istanbul anonymous names join to the unique function at the same file and line", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "crap4ts-coverage-"));
  const coveragePath = path.join(dir, "coverage-final.json");
  fs.writeFileSync(
    coveragePath,
    JSON.stringify({
      [path.join(dir, "src/example.ts")]: {
        fnMap: {
          "0": {
            name: "(anonymous_0)",
            decl: { start: { line: 2 }, end: { line: 2 } },
            loc: { start: { line: 2 }, end: { line: 4 } },
          },
        },
        f: { "0": 1 },
        statementMap: {
          "0": { start: { line: 3 }, end: { line: 3 } },
          "1": { start: { line: 4 }, end: { line: 4 } },
        },
        s: { "0": 1, "1": 0 },
      },
    }),
    "utf8",
  );

  const coverage = readCoverage(coveragePath, dir, [method]);

  assert.equal(coverage.get("src/example.ts:2:Example.method"), 0.5);
});

test("lcov anonymous names join to the unique function at the same file and line", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "crap4ts-lcov-"));
  const coveragePath = path.join(dir, "lcov.info");
  fs.writeFileSync(
    coveragePath,
    `SF:${path.join(dir, "src/example.ts")}
FN:2,(anonymous_0)
FNDA:1,(anonymous_0)
end_of_record
`,
    "utf8",
  );

  const coverage = readCoverage(coveragePath, dir, [method]);

  assert.equal(coverage.get("src/example.ts:2:Example.method"), 1);
});
