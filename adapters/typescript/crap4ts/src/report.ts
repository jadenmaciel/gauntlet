import type { CliReport, Report } from "./types.js";

export function toCliReport(report: Report): CliReport {
  return {
    ceiling: report.ceiling,
    functions: report.functions.map((fn) => ({
      file: fn.file,
      line: fn.line,
      func: fn.func,
      complexity: fn.complexity,
      coverage: fn.coverage,
      crap: fn.crap,
      pass: fn.pass,
    })),
    summary: report.summary,
  };
}

export function formatText(report: Report): string {
  const lines: string[] = [];
  for (const fn of report.functions) {
    const status = fn.pass ? "PASS" : "FAIL";
    const note = fn.matched ? "" : " (no coverage data)";
    lines.push(
      `${fn.file}:${fn.line}:${fn.func}\tcomplexity=${fn.complexity}\tcoverage=${(fn.coverage * 100).toFixed(1)}%\tcrap=${fn.crap.toFixed(2)}\t${status}${note}`,
    );
  }
  lines.push("");
  lines.push(
    `ceiling=${report.ceiling.toFixed(2)} total=${report.summary.total} failing=${report.summary.failing} max_crap=${report.summary.max_crap.toFixed(2)}`,
  );
  return lines.join("\n");
}

export function reportExitCode(report: Report): number {
  return report.summary.failing > 0 ? 1 : 0;
}
