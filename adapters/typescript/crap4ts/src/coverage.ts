import fs from "node:fs";
import path from "node:path";

import { buildKey } from "./key.js";
import type { FunctionComplexity } from "./types.js";

interface Position {
  line: number;
  column?: number;
}

interface SourceRange {
  start: Position;
  end: Position;
}

interface IstanbulFileCoverage {
  fnMap?: Record<string, { name: string; decl: SourceRange; loc: SourceRange }>;
  f?: Record<string, number>;
  statementMap?: Record<string, SourceRange>;
  s?: Record<string, number>;
}

export function readCoverage(
  coveragePath: string,
  rootDir: string,
  functions: FunctionComplexity[],
): Map<string, number> {
  const functionNames = indexFunctionNames(functions);
  const ext = path.extname(coveragePath).toLowerCase();
  if (ext === ".json") {
    return readCoverageFinalJSON(coveragePath, rootDir, functionNames);
  }
  if (ext === ".lcov" || ext === ".info") {
    return readLcov(coveragePath, rootDir, functionNames);
  }
  throw new Error(`unsupported coverage format for ${coveragePath}; expected .json, .lcov, or .info`);
}

function readCoverageFinalJSON(
  coveragePath: string,
  rootDir: string,
  functionNames: Map<string, string[]>,
): Map<string, number> {
  const raw = fs.readFileSync(coveragePath, "utf8");
  const decoded: unknown = JSON.parse(raw);
  if (!decoded || typeof decoded !== "object") {
    throw new Error(`invalid coverage JSON: ${coveragePath}`);
  }

  const coverageMap = new Map<string, number>();
  const entries = Object.entries(decoded as Record<string, IstanbulFileCoverage>);
  for (const [rawFile, fileCoverage] of entries) {
    const fnMap = fileCoverage.fnMap ?? {};
    const fnHits = fileCoverage.f ?? {};
    const statementMap = fileCoverage.statementMap ?? {};
    const statementHits = fileCoverage.s ?? {};
    const normalizedFile = normalizeFile(rawFile, rootDir);

    for (const [fnID, fn] of Object.entries(fnMap)) {
      const line = fn.decl?.start?.line;
      if (!line || !Number.isFinite(line)) {
        continue;
      }
      const key = coverageKey(normalizedFile, line, fn.name, functionNames);
      const nestedRanges = Object.entries(fnMap)
        .filter(([candidateID, candidate]) => candidateID !== fnID && rangeStrictlyContains(fn.loc, candidate.loc))
        .map(([, candidate]) => candidate.loc);
      const coverage = coverageFromStatements(fn.loc, nestedRanges, statementMap, statementHits, fnHits[fnID]);
      coverageMap.set(key, clampCoverage(coverage));
    }
  }
  return coverageMap;
}

function coverageFromStatements(
  fnRange: SourceRange,
  nestedRanges: SourceRange[],
  statementMap: Record<string, SourceRange>,
  statementHits: Record<string, number>,
  fnHitCount: number | undefined,
): number {
  const startLine = fnRange?.start?.line ?? 0;
  const endLine = fnRange?.end?.line ?? 0;
  if (startLine <= 0 || endLine <= 0 || endLine < startLine) {
    return fnHitCount && fnHitCount > 0 ? 1 : 0;
  }

  const inRangeStatementIDs: string[] = [];
  for (const [statementID, statementRange] of Object.entries(statementMap)) {
    const statementLine = statementRange.start?.line ?? 0;
    const belongsToNestedFunction = nestedRanges.some((nestedRange) => rangeContainsPosition(nestedRange, statementRange.start));
    if (statementLine >= startLine && statementLine <= endLine && !belongsToNestedFunction) {
      inRangeStatementIDs.push(statementID);
    }
  }
  if (inRangeStatementIDs.length === 0) {
    return fnHitCount && fnHitCount > 0 ? 1 : 0;
  }

  let covered = 0;
  for (const statementID of inRangeStatementIDs) {
    const hits = statementHits[statementID] ?? 0;
    if (hits > 0) {
      covered += 1;
    }
  }
  return covered / inRangeStatementIDs.length;
}

function rangeStrictlyContains(outer: SourceRange, inner: SourceRange): boolean {
  return (
    comparePositions(outer.start, inner.start) <= 0 &&
    comparePositions(outer.end, inner.end) >= 0 &&
    (comparePositions(outer.start, inner.start) < 0 || comparePositions(outer.end, inner.end) > 0)
  );
}

function rangeContainsPosition(range: SourceRange, position: Position): boolean {
  return comparePositions(range.start, position) <= 0 && comparePositions(range.end, position) >= 0;
}

function comparePositions(left: Position, right: Position): number {
  if (left.line !== right.line) {
    return left.line - right.line;
  }
  return (left.column ?? 0) - (right.column ?? 0);
}

function readLcov(
  coveragePath: string,
  rootDir: string,
  functionNames: Map<string, string[]>,
): Map<string, number> {
  const raw = fs.readFileSync(coveragePath, "utf8");
  const coverageMap = new Map<string, number>();

  let currentFile = "";
  const functionLines = new Map<string, number>();
  const lines = raw.split(/\r?\n/);
  for (const line of lines) {
    if (line.startsWith("SF:")) {
      currentFile = normalizeFile(line.slice(3).trim(), rootDir);
      functionLines.clear();
      continue;
    }
    if (line.startsWith("FN:")) {
      const commaIdx = line.indexOf(",");
      if (commaIdx === -1) {
        continue;
      }
      const lineNumber = Number(line.slice(3, commaIdx));
      const fnName = line.slice(commaIdx + 1).trim();
      if (Number.isFinite(lineNumber) && fnName.length > 0) {
        functionLines.set(fnName, lineNumber);
      }
      continue;
    }
    if (line.startsWith("FNDA:")) {
      const commaIdx = line.indexOf(",");
      if (commaIdx === -1 || !currentFile) {
        continue;
      }
      const hits = Number(line.slice(5, commaIdx));
      const fnName = line.slice(commaIdx + 1).trim();
      const fnLine = functionLines.get(fnName);
      if (!fnLine) {
        continue;
      }
      const key = coverageKey(currentFile, fnLine, fnName, functionNames);
      coverageMap.set(key, hits > 0 ? 1 : 0);
    }
  }

  return coverageMap;
}

function indexFunctionNames(functions: FunctionComplexity[]): Map<string, string[]> {
  const names = new Map<string, string[]>();
  for (const fn of functions) {
    const key = locationKey(fn.file, fn.line);
    const atLocation = names.get(key) ?? [];
    atLocation.push(fn.func);
    names.set(key, atLocation);
  }
  return names;
}

function coverageKey(
  file: string,
  line: number,
  coverageName: string,
  functionNames: Map<string, string[]>,
): string {
  const names = functionNames.get(locationKey(file, line)) ?? [];
  if (names.includes(coverageName) || names.length !== 1) {
    return buildKey(file, line, coverageName);
  }
  return buildKey(file, line, names[0]);
}

function locationKey(file: string, line: number): string {
  return `${file}\0${line}`;
}

function normalizeFile(filePath: string, rootDir: string): string {
  const abs = path.isAbsolute(filePath) ? filePath : path.resolve(rootDir, filePath);
  const rel = path.relative(rootDir, abs);
  return rel.split(path.sep).join("/");
}

function clampCoverage(value: number): number {
  if (!Number.isFinite(value)) {
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
