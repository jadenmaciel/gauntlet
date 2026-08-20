import fs from "node:fs";
import path from "node:path";

import { buildKey } from "./key.js";

interface Position {
  line: number;
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

export function readCoverage(coveragePath: string, rootDir: string): Map<string, number> {
  const ext = path.extname(coveragePath).toLowerCase();
  if (ext === ".json") {
    return readCoverageFinalJSON(coveragePath, rootDir);
  }
  if (ext === ".lcov" || ext === ".info") {
    return readLcov(coveragePath, rootDir);
  }
  throw new Error(`unsupported coverage format for ${coveragePath}; expected .json, .lcov, or .info`);
}

function readCoverageFinalJSON(coveragePath: string, rootDir: string): Map<string, number> {
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
      const key = buildKey(normalizedFile, line, fn.name);
      const coverage = coverageFromStatements(fn.loc, statementMap, statementHits, fnHits[fnID]);
      coverageMap.set(key, clampCoverage(coverage));
    }
  }
  return coverageMap;
}

function coverageFromStatements(
  fnRange: SourceRange,
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
    if (statementLine >= startLine && statementLine <= endLine) {
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

function readLcov(coveragePath: string, rootDir: string): Map<string, number> {
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
      const key = buildKey(currentFile, fnLine, fnName);
      coverageMap.set(key, hits > 0 ? 1 : 0);
    }
  }

  return coverageMap;
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
