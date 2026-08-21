#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import { parseArgs } from "node:util";
import type { Writable } from "node:stream";

import { scanComplexity } from "./complexity.js";
import { readCoverage } from "./coverage.js";
import { evaluate } from "./evaluate.js";
import { formatText, reportExitCode, toCliReport } from "./report.js";
import { readCeiling } from "./thresholds.js";

interface CliOptions {
  dir: string;
  coverage: string;
  thresholds: string;
  format: "text" | "json";
  ceiling: number | null;
  changed: string;
}

const USAGE = `crap4ts: score TypeScript functions with CRAP

Usage:
  crap4ts [--dir <path>] [--coverage <coverage-final.json|lcov.info>] [--thresholds <.gauntlet/thresholds.yml>] [--ceiling <number>] [--changed <ref>] [--format <text|json>]

Examples:
  crap4ts --dir . --coverage coverage/coverage-final.json --format json
  crap4ts --dir . --coverage coverage/lcov.info --ceiling 30
  crap4ts --dir . --coverage coverage/coverage-final.json --changed origin/main
`;

export async function run(args: string[], stdout: Writable, stderr: Writable): Promise<number> {
  let options: CliOptions;
  try {
    options = parseOptions(args);
  } catch (error) {
    stderr.write(`crap4ts: ${(error as Error).message}\n`);
    return 2;
  }

  try {
    let functions = scanComplexity(options.dir);
    if (options.changed !== "") {
      const changedFiles = gitChangedFiles(options.dir, options.changed);
      if (changedFiles.length === 0) {
        stderr.write(`crap4ts: no files changed since ${JSON.stringify(options.changed)}; nothing was scored\n`);
      }
      functions = functions.filter((fn) => changedFiles.some((candidate) => filesMatch(fn.file, candidate)));
    }
    const coverage = readCoverage(options.coverage, options.dir, functions);
    const ceiling = options.ceiling ?? readCeiling(options.thresholds);
    const report = evaluate(functions, coverage, ceiling);

    if (options.format === "json") {
      stdout.write(`${JSON.stringify(toCliReport(report), null, 2)}\n`);
    } else {
      stdout.write(`${formatText(report)}\n`);
    }
    return reportExitCode(report);
  } catch (error) {
    stderr.write(`crap4ts: ${(error as Error).message}\n`);
    return 2;
  }
}

function parseOptions(args: string[]): CliOptions {
  const parsed = parseArgs({
    args,
    options: {
      dir: { type: "string", default: "." },
      coverage: { type: "string", default: "coverage/coverage-final.json" },
      thresholds: { type: "string", default: ".gauntlet/thresholds.yml" },
      format: { type: "string", default: "text" },
      ceiling: { type: "string" },
      changed: { type: "string", default: "" },
      help: { type: "boolean", short: "h" },
    },
    allowPositionals: false,
    strict: true,
  });

  if (parsed.values.help) {
    throw new Error(USAGE);
  }

  if (parsed.values.format !== "text" && parsed.values.format !== "json") {
    throw new Error(`invalid --format ${JSON.stringify(parsed.values.format)}, want "text" or "json"`);
  }

  let ceiling: number | null = null;
  if (parsed.values.ceiling !== undefined) {
    const decoded = Number(parsed.values.ceiling);
    if (!Number.isFinite(decoded)) {
      throw new Error(`invalid --ceiling ${JSON.stringify(parsed.values.ceiling)}`);
    }
    ceiling = decoded;
  }

  return {
    dir: parsed.values.dir,
    coverage: parsed.values.coverage,
    thresholds: parsed.values.thresholds,
    format: parsed.values.format,
    ceiling,
    changed: parsed.values.changed,
  };
}

/**
 * Files touched since the merge base of `ref` and HEAD, plus untracked ones.
 */
export function gitChangedFiles(dir: string, ref: string): string[] {
  const base = git(dir, ["merge-base", ref, "HEAD"]).trim();
  const diff = git(dir, ["diff", "--name-only", base]);
  const untracked = git(dir, ["ls-files", "--others", "--exclude-standard"]);
  return `${diff}\n${untracked}`
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "");
}

function git(dir: string, args: string[]): string {
  try {
    return execFileSync("git", args, { cwd: dir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] });
  } catch (error) {
    const stderr = (error as { stderr?: string }).stderr ?? "";
    throw new Error(`git ${args.join(" ")}: ${stderr.trim() || (error as Error).message}`);
  }
}

/**
 * Compare two paths allowing a path-segment suffix match either way, because
 * `git diff --name-only` reports paths from the repo root while scanned files
 * are relative to --dir, which may sit below it.
 */
export function filesMatch(left: string, right: string): boolean {
  const a = left.split("\\").join("/");
  const b = right.split("\\").join("/");
  return a === b || a.endsWith(`/${b}`) || b.endsWith(`/${a}`);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  run(process.argv.slice(2), process.stdout, process.stderr).then((code) => {
    process.exit(code);
  });
}
