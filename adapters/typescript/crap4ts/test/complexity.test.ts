import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { scanComplexity } from "../src/complexity.js";

test("TypeScript declarations use Istanbul-compatible lines and standard decision points", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "crap4ts-complexity-"));
  const srcDir = path.join(dir, "src");
  fs.mkdirSync(srcDir);
  fs.writeFileSync(
    path.join(srcDir, "example.ts"),
    `class Example {
  @logged
  method(input: boolean): number {
    const nested = () => input && true;
    if (input) return 1;
    return nested() ? 2 : 0;
  }
}

export function decisions(items: number[]): number {
  let total = 0;
  try {
    for (const item of items) {
      if (item > 0 && item < 10) {
        total += item > 5 ? 2 : 1;
      }
    }
  } catch {
    return -1;
  }
  switch (total) {
    case 1:
      return 1;
    case 2:
      return 2;
    default:
      return total ?? 0;
  }
}
`,
    "utf8",
  );

  const functions = scanComplexity(dir);
  const byName = new Map(functions.map((fn) => [fn.func, fn]));

  assert.deepEqual(byName.get("Example.method"), {
    file: "src/example.ts",
    line: 3,
    func: "Example.method",
    complexity: 3,
  });
  assert.deepEqual(byName.get("nested"), {
    file: "src/example.ts",
    line: 4,
    func: "nested",
    complexity: 2,
  });
  assert.deepEqual(byName.get("decisions"), {
    file: "src/example.ts",
    line: 10,
    func: "decisions",
    complexity: 9,
  });
});

test("invalid TypeScript fails the complexity scan", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "crap4ts-invalid-"));
  fs.writeFileSync(path.join(dir, "invalid.ts"), "export function broken( {", "utf8");

  assert.throws(() => scanComplexity(dir), /invalid\.ts:1:/);
});

test("test source files are excluded from complexity scans", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "crap4ts-tests-"));
  const testsDir = path.join(dir, "__tests__");
  fs.mkdirSync(testsDir);
  fs.writeFileSync(path.join(dir, "source.ts"), "export function source() { return 1; }", "utf8");
  fs.writeFileSync(path.join(dir, "source.test.ts"), "export function unitTest() { return 1; }", "utf8");
  fs.writeFileSync(path.join(dir, "source.spec.tsx"), "export function spec() { return 1; }", "utf8");
  fs.writeFileSync(path.join(testsDir, "helper.ts"), "export function helper() { return 1; }", "utf8");

  const functions = scanComplexity(dir);

  assert.deepEqual(
    functions.map((fn) => fn.func),
    ["source"],
  );
});
