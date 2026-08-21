import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { Writable } from "node:stream";
import test from "node:test";

import { filesMatch, run } from "../src/cli.js";

class StringSink extends Writable {
  private readonly chunks: string[] = [];

  override _write(chunk: Buffer | string, _encoding: BufferEncoding, callback: (error?: Error | null) => void): void {
    this.chunks.push(typeof chunk === "string" ? chunk : chunk.toString("utf8"));
    callback();
  }

  text(): string {
    return this.chunks.join("");
  }
}

test("cli prints required json report shape and exits 1 for failing functions", async () => {
  const fixture = writeFixtureProject();
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    [
      "--dir",
      fixture.dir,
      "--coverage",
      fixture.lowCoveragePath,
      "--thresholds",
      fixture.thresholdsPath,
      "--format",
      "json",
    ],
    stdout,
    stderr,
  );

  assert.equal(stderr.text(), "");
  assert.equal(code, 1);

  const decoded = JSON.parse(stdout.text()) as {
    ceiling: number;
    functions: Array<{
      file: string;
      line: number;
      func: string;
      complexity: number;
      coverage: number;
      crap: number;
      pass: boolean;
    }>;
    summary: { total: number; failing: number; max_crap: number };
  };
  assert.equal(decoded.ceiling, 30);
  assert.equal(decoded.summary.total, 1);
  assert.equal(decoded.summary.failing, 1);
  assert.equal(decoded.functions.length, 1);
  assert.equal(decoded.functions[0].file, "src/hot.ts");
  assert.equal(decoded.functions[0].func, "hotPath");
  assert.equal(decoded.functions[0].coverage, 10);
});

test("cli exits 0 when coverage increase drops CRAP below ceiling", async () => {
  const fixture = writeFixtureProject();
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    [
      "--dir",
      fixture.dir,
      "--coverage",
      fixture.highCoveragePath,
      "--thresholds",
      fixture.thresholdsPath,
      "--format",
      "json",
    ],
    stdout,
    stderr,
  );

  assert.equal(stderr.text(), "");
  assert.equal(code, 0);
  const decoded = JSON.parse(stdout.text()) as {
    functions: Array<{ coverage: number }>;
    summary: { failing: number };
  };
  assert.equal(decoded.summary.failing, 0);
  assert.equal(decoded.functions[0].coverage, 100);
});

test("cli honors --ceiling override", async () => {
  const fixture = writeFixtureProject();
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    [
      "--dir",
      fixture.dir,
      "--coverage",
      fixture.lowCoveragePath,
      "--thresholds",
      fixture.thresholdsPath,
      "--ceiling",
      "100",
      "--format",
      "json",
    ],
    stdout,
    stderr,
  );

  assert.equal(stderr.text(), "");
  assert.equal(code, 0);
  const decoded = JSON.parse(stdout.text()) as { ceiling: number; summary: { failing: number } };
  assert.equal(decoded.ceiling, 100);
  assert.equal(decoded.summary.failing, 0);
});

test("cli scopes the report to files changed since --changed", async () => {
  const fixture = writeGitFixtureProject(false);
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    ["--dir", fixture.dir, "--coverage", fixture.lowCoveragePath, "--ceiling", "100", "--changed", "HEAD", "--format", "json"],
    stdout,
    stderr,
  );

  assert.equal(code, 0, stderr.text());
  const decoded = JSON.parse(stdout.text()) as { functions: { file: string }[] };
  assert.deepEqual(
    decoded.functions.map((fn) => fn.file),
    ["src/touched.ts"],
  );
});

test("cli warns instead of passing silently when --changed finds an empty diff", async () => {
  const fixture = writeGitFixtureProject(true);
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    ["--dir", fixture.dir, "--coverage", fixture.lowCoveragePath, "--ceiling", "1", "--changed", "HEAD", "--format", "json"],
    stdout,
    stderr,
  );

  assert.equal(code, 0, stderr.text());
  assert.match(stderr.text(), /no files changed/);
  const decoded = JSON.parse(stdout.text()) as { summary: { total: number } };
  assert.equal(decoded.summary.total, 0);
});

test("cli reports an unknown --changed ref as a usage error", async () => {
  const fixture = writeGitFixtureProject(false);
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    ["--dir", fixture.dir, "--coverage", fixture.lowCoveragePath, "--ceiling", "100", "--changed", "no-such-ref"],
    stdout,
    stderr,
  );

  assert.equal(code, 2);
  assert.match(stderr.text(), /merge-base/);
});

test("cli reports a missing coverage file as a usage error", async () => {
  const fixture = writeFixtureProject();
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    ["--dir", fixture.dir, "--coverage", path.join(fixture.dir, "absent.json"), "--ceiling", "100"],
    stdout,
    stderr,
  );

  assert.equal(code, 2);
});

test("cli reports a missing ceiling source as a usage error", async () => {
  const fixture = writeFixtureProject();
  const stdout = new StringSink();
  const stderr = new StringSink();

  const code = await run(
    ["--dir", fixture.dir, "--coverage", fixture.lowCoveragePath, "--thresholds", path.join(fixture.dir, "absent.yml")],
    stdout,
    stderr,
  );

  assert.equal(code, 2);
});

test("filesMatch allows a path-segment suffix in either direction", () => {
  assert.equal(filesMatch("src/hot.ts", "src/hot.ts"), true);
  assert.equal(filesMatch("hot.ts", "src/hot.ts"), true);
  assert.equal(filesMatch("src/hot.ts", "hot.ts"), true);
  assert.equal(filesMatch("nothot.ts", "hot.ts"), false);
  assert.equal(filesMatch("src/a.ts", "src/b.ts"), false);
});

function writeGitFixtureProject(commitEverything: boolean): FixtureProject {
  const fixture = writeFixtureProject();
  git(fixture.dir, ["init", "-q"]);
  git(fixture.dir, ["add", "."]);
  commit(fixture.dir, "initial");
  fs.writeFileSync(
    path.join(fixture.dir, "src/touched.ts"),
    "export function touched(input: number): number {\n  if (input > 0) return 1;\n  return 0;\n}\n",
    "utf8",
  );
  if (commitEverything) {
    git(fixture.dir, ["add", "."]);
    commit(fixture.dir, "second");
  }
  return fixture;
}

function commit(dir: string, message: string): void {
  git(dir, ["-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-qm", message]);
}

function git(dir: string, args: string[]): void {
  // Ignore the developer's global and system git config. A global
  // core.hooksPath can drop generated files into the fixture on commit, which
  // then show up as untracked changes and skew the assertions.
  execFileSync("git", args, {
    cwd: dir,
    env: { ...process.env, GIT_CONFIG_GLOBAL: "/dev/null", GIT_CONFIG_SYSTEM: "/dev/null" },
    stdio: "ignore",
  });
}

interface FixtureProject {
  dir: string;
  thresholdsPath: string;
  lowCoveragePath: string;
  highCoveragePath: string;
}

function writeFixtureProject(): FixtureProject {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "crap4ts-fixture-"));
  const srcDir = path.join(dir, "src");
  const gauntletDir = path.join(dir, ".gauntlet");
  fs.mkdirSync(srcDir, { recursive: true });
  fs.mkdirSync(gauntletDir, { recursive: true });

  const sourcePath = path.join(srcDir, "hot.ts");
  fs.writeFileSync(
    sourcePath,
    `export function hotPath(input: number): number {
  if (input > 10) return 1;
  if (input > 9) return 2;
  if (input > 8) return 3;
  if (input > 7) return 4;
  if (input > 6) return 5;
  if (input > 5) return 6;
  if (input > 4) return 7;
  if (input > 3) return 8;
  if (input > 2) return 9;
  return 10;
}
`,
    "utf8",
  );

  const thresholdsPath = path.join(gauntletDir, "thresholds.yml");
  fs.writeFileSync(
    thresholdsPath,
    `metrics:
  crap_ceiling:
    direction: max
    value: 30
`,
    "utf8",
  );

  const lowCoveragePath = path.join(dir, "coverage-low.json");
  fs.writeFileSync(lowCoveragePath, JSON.stringify(coverageDocument(sourcePath, [1, 0, 0, 0, 0, 0, 0, 0, 0, 0])), "utf8");

  const highCoveragePath = path.join(dir, "coverage-high.json");
  fs.writeFileSync(highCoveragePath, JSON.stringify(coverageDocument(sourcePath, [1, 1, 1, 1, 1, 1, 1, 1, 1, 1])), "utf8");

  return { dir, thresholdsPath, lowCoveragePath, highCoveragePath };
}

function coverageDocument(filePath: string, statementHits: number[]): Record<string, unknown> {
  return {
    [filePath]: {
      path: filePath,
      statementMap: {
        "0": { start: { line: 2, column: 2 }, end: { line: 2, column: 26 } },
        "1": { start: { line: 3, column: 2 }, end: { line: 3, column: 25 } },
        "2": { start: { line: 4, column: 2 }, end: { line: 4, column: 25 } },
        "3": { start: { line: 5, column: 2 }, end: { line: 5, column: 25 } },
        "4": { start: { line: 6, column: 2 }, end: { line: 6, column: 25 } },
        "5": { start: { line: 7, column: 2 }, end: { line: 7, column: 25 } },
        "6": { start: { line: 8, column: 2 }, end: { line: 8, column: 25 } },
        "7": { start: { line: 9, column: 2 }, end: { line: 9, column: 25 } },
        "8": { start: { line: 10, column: 2 }, end: { line: 10, column: 25 } },
        "9": { start: { line: 11, column: 2 }, end: { line: 11, column: 12 } }
      },
      s: {
        "0": statementHits[0],
        "1": statementHits[1],
        "2": statementHits[2],
        "3": statementHits[3],
        "4": statementHits[4],
        "5": statementHits[5],
        "6": statementHits[6],
        "7": statementHits[7],
        "8": statementHits[8],
        "9": statementHits[9]
      },
      fnMap: {
        "0": {
          name: "hotPath",
          decl: { start: { line: 1, column: 15 }, end: { line: 1, column: 22 } },
          loc: { start: { line: 1, column: 0 }, end: { line: 12, column: 1 } }
        }
      },
      f: {
        "0": statementHits[0]
      }
    }
  };
}
