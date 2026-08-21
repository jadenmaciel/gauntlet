#!/usr/bin/env python3
from __future__ import annotations

import argparse
import ast
import json
import os
import subprocess
import xml.etree.ElementTree as ET
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, Sequence, TextIO


@dataclass(frozen=True)
class FunctionComplexity:
    file: str
    line: int
    func: str
    complexity: int
    coverage_lines: tuple[int, ...] = ()


@dataclass(frozen=True)
class CoverageEntry:
    coverage: float


@dataclass(frozen=True)
class FunctionReport:
    file: str
    line: int
    func: str
    complexity: int
    coverage: float
    crap: float
    pass_: bool
    matched: bool


@dataclass(frozen=True)
class Summary:
    total: int
    failing: int
    max_crap: float


@dataclass(frozen=True)
class Report:
    ceiling: float
    functions: list[FunctionReport]
    summary: Summary


def build_key(file: str, line: int, name: str) -> str:
    return f"{file}:{line}:{name}"


def compute_crap(complexity: int, coverage_percent: float) -> float:
    cc = float(complexity)
    coverage = coverage_percent / 100.0
    uncovered = 1.0 - coverage
    return cc * cc * uncovered * uncovered * uncovered + cc


class _ComplexityCounter(ast.NodeVisitor):
    def __init__(self) -> None:
        self.value = 1

    def visit_If(self, node: ast.If) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_For(self, node: ast.For) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_AsyncFor(self, node: ast.AsyncFor) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_While(self, node: ast.While) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_ExceptHandler(self, node: ast.ExceptHandler) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_Assert(self, node: ast.Assert) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_IfExp(self, node: ast.IfExp) -> None:
        self.value += 1
        self.generic_visit(node)

    def visit_BoolOp(self, node: ast.BoolOp) -> None:
        self.value += max(0, len(node.values) - 1)
        self.generic_visit(node)

    def visit_comprehension(self, node: ast.comprehension) -> None:
        self.value += 1 + len(node.ifs)
        self.generic_visit(node)

    def visit_Match(self, node: ast.Match) -> None:
        for case in node.cases:
            if not _is_wildcard_match_case(case):
                self.value += 1
        self.generic_visit(node)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        return

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        return

    def visit_Lambda(self, node: ast.Lambda) -> None:
        return


def _is_wildcard_match_case(case: ast.match_case) -> bool:
    return (
        isinstance(case.pattern, ast.MatchAs)
        and case.pattern.name is None
        and case.guard is None
    )


def function_complexity(node: ast.FunctionDef | ast.AsyncFunctionDef) -> int:
    counter = _ComplexityCounter()
    for statement in node.body:
        counter.visit(statement)
    return counter.value


def _coverage_lines(node: ast.FunctionDef | ast.AsyncFunctionDef) -> tuple[int, ...]:
    if not node.body or node.end_lineno is None:
        return ()
    lines = set(range(node.body[0].lineno, node.end_lineno + 1))
    for child in ast.walk(node):
        if child is node or not isinstance(child, (ast.FunctionDef, ast.AsyncFunctionDef)):
            continue
        child_end = child.end_lineno or child.lineno
        lines.difference_update(range(child.lineno, child_end + 1))
    return tuple(sorted(lines))


class _FunctionCollector(ast.NodeVisitor):
    def __init__(self, file_path: str) -> None:
        self._file_path = file_path
        self.functions: list[FunctionComplexity] = []

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        self.functions.append(
            FunctionComplexity(
                file=self._file_path,
                line=node.lineno,
                func=node.name,
                complexity=function_complexity(node),
                coverage_lines=_coverage_lines(node),
            )
        )
        self.generic_visit(node)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        self.functions.append(
            FunctionComplexity(
                file=self._file_path,
                line=node.lineno,
                func=node.name,
                complexity=function_complexity(node),
                coverage_lines=_coverage_lines(node),
            )
        )
        self.generic_visit(node)


def _normalize_file_path(file_path: str, root_dir: str) -> str:
    normalized = file_path.replace("\\", "/")
    root = Path(root_dir).resolve()
    candidate = Path(normalized)

    if candidate.is_absolute():
        try:
            normalized = candidate.resolve().relative_to(root).as_posix()
        except ValueError:
            normalized = candidate.as_posix()
    else:
        normalized = Path(normalized).as_posix()
        if normalized.startswith("./"):
            normalized = normalized[2:]
    return normalized


def scan_python_file(path: Path, root_dir: str) -> list[FunctionComplexity]:
    source = path.read_text(encoding="utf-8")
    tree = ast.parse(source, filename=str(path))
    collector = _FunctionCollector(_normalize_file_path(str(path), root_dir))
    collector.visit(tree)
    return collector.functions


def scan_dir(scan_dir: str) -> list[FunctionComplexity]:
    base = Path(scan_dir).resolve()
    excluded_dirs = {
        ".git",
        "__pycache__",
        ".mypy_cache",
        ".pytest_cache",
        ".ruff_cache",
        ".venv",
        "venv",
    }
    functions: list[FunctionComplexity] = []
    for root, dirs, files in os.walk(base):
        dirs[:] = [name for name in dirs if name not in excluded_dirs]
        for file_name in files:
            if not file_name.endswith(".py"):
                continue
            file_path = Path(root) / file_name
            functions.extend(scan_python_file(file_path, str(base)))
    functions.sort(key=lambda fn: (fn.file, fn.line, fn.func))
    return functions


def _coverage_file_path(
    filename: str, root_dir: str, sources: Sequence[str]
) -> str:
    candidate = Path(filename)
    if candidate.is_absolute():
        return _normalize_file_path(filename, root_dir)

    root = Path(root_dir).resolve()
    for source in sources:
        try:
            return (Path(source).resolve() / candidate).relative_to(root).as_posix()
        except ValueError:
            continue
    return _normalize_file_path(filename, root_dir)


def load_coverage_xml(
    path: str, root_dir: str, functions: Iterable[FunctionComplexity]
) -> dict[str, CoverageEntry]:
    tree = ET.parse(path)
    root = tree.getroot()
    sources = [
        source.text.strip()
        for source in root.findall("./sources/source")
        if source.text and source.text.strip()
    ]
    line_hits_by_file: dict[str, dict[int, int]] = {}

    for class_node in root.findall(".//class"):
        filename = class_node.get("filename")
        if not filename:
            continue
        normalized_file = _coverage_file_path(filename, root_dir, sources)
        line_hits = line_hits_by_file.setdefault(normalized_file, {})
        for line_node in class_node.findall("./lines/line"):
            line_value = line_node.get("number")
            if not line_value:
                continue
            line = int(line_value)
            line_hits[line] = max(line_hits.get(line, 0), int(line_node.get("hits", "0")))

    coverage: dict[str, CoverageEntry] = {}
    for function in functions:
        file_hits = line_hits_by_file.get(function.file)
        if file_hits is None:
            continue
        hits = [file_hits[line] for line in function.coverage_lines if line in file_hits]
        if not hits:
            continue
        coverage_percent = sum(hit > 0 for hit in hits) / len(hits) * 100.0
        coverage[build_key(function.file, function.line, function.func)] = CoverageEntry(
            coverage=coverage_percent
        )

    return coverage


def _git_output(root_dir: str, args: Sequence[str]) -> str:
    result = subprocess.run(
        ["git", *args], cwd=root_dir, capture_output=True, text=True, check=False
    )
    if result.returncode != 0:
        raise ValueError(f"git {' '.join(args)}: {result.stderr.strip()}")
    return result.stdout


def git_changed_files(root_dir: str, ref: str) -> list[str]:
    """Files touched since the merge base of ref and HEAD, plus untracked ones."""
    base = _git_output(root_dir, ["merge-base", ref, "HEAD"]).strip()
    diff = _git_output(root_dir, ["diff", "--name-only", base])
    untracked = _git_output(root_dir, ["ls-files", "--others", "--exclude-standard"])
    return [line.strip() for line in (diff + untracked).splitlines() if line.strip()]


def files_match(a: str, b: str) -> bool:
    """Compare two paths allowing a path-segment suffix match either way.

    `git diff --name-only` reports paths from the repo root, while scanned
    files are relative to --dir, which may sit below it.
    """
    a = a.replace("\\", "/")
    b = b.replace("\\", "/")
    return a == b or a.endswith("/" + b) or b.endswith("/" + a)


def filter_by_changed_files(
    functions: Iterable[FunctionComplexity], changed_files: Sequence[str]
) -> list[FunctionComplexity]:
    return [
        fn for fn in functions if any(files_match(fn.file, cf) for cf in changed_files)
    ]


def evaluate(
    functions: Iterable[FunctionComplexity], coverage: dict[str, CoverageEntry], ceiling: float
) -> Report:
    report_functions: list[FunctionReport] = []
    failing = 0
    max_crap = 0.0

    for fn in functions:
        entry = coverage.get(build_key(fn.file, fn.line, fn.func))
        matched = entry is not None
        coverage_percent = entry.coverage if entry else 0.0
        score = compute_crap(fn.complexity, coverage_percent)
        passed = score <= ceiling
        if not passed:
            failing += 1
        max_crap = max(max_crap, score)
        report_functions.append(
            FunctionReport(
                file=fn.file,
                line=fn.line,
                func=fn.func,
                complexity=fn.complexity,
                coverage=coverage_percent,
                crap=score,
                pass_=passed,
                matched=matched,
            )
        )

    return Report(
        ceiling=ceiling,
        functions=report_functions,
        summary=Summary(total=len(report_functions), failing=failing, max_crap=max_crap),
    )


def read_ceiling_from_thresholds(path: str) -> float:
    lines = Path(path).read_text(encoding="utf-8").splitlines()
    metrics_indent: int | None = None
    crap_indent: int | None = None

    for raw_line in lines:
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue

        indent = len(raw_line) - len(raw_line.lstrip(" "))
        token = raw_line.strip()

        if token == "metrics:":
            metrics_indent = indent
            crap_indent = None
            continue

        if metrics_indent is None:
            continue

        if indent <= metrics_indent:
            metrics_indent = None
            crap_indent = None
            continue

        if crap_indent is None:
            if indent == metrics_indent + 2 and token == "crap_ceiling:":
                crap_indent = indent
            continue

        if indent <= crap_indent:
            crap_indent = None
            if indent == metrics_indent + 2 and token == "crap_ceiling:":
                crap_indent = indent
            continue

        if indent >= crap_indent + 2 and token.startswith("value:"):
            raw_value = token.split(":", 1)[1].strip()
            return float(raw_value)

    raise ValueError(f"unable to read metrics.crap_ceiling.value from {path}")


class _ParserExit(Exception):
    def __init__(self, status: int) -> None:
        super().__init__(status)
        self.status = status


class _ArgumentParser(argparse.ArgumentParser):
    def __init__(self, stderr: TextIO) -> None:
        super().__init__(prog="crap4py")
        self._stderr = stderr

    def _print_message(self, message: str | None, file: TextIO | None = None) -> None:
        if message:
            self._stderr.write(message)

    def exit(self, status: int = 0, message: str | None = None) -> None:
        if message:
            self._stderr.write(message)
        raise _ParserExit(status)


def parse_options(argv: Sequence[str], stderr: TextIO) -> argparse.Namespace:
    parser = _ArgumentParser(stderr)
    parser.add_argument("--dir", default=".", help="directory to scan for .py source files")
    parser.add_argument(
        "--coverage",
        default="coverage.xml",
        help="coverage.py XML report path (generate with `coverage xml`)",
    )
    parser.add_argument(
        "--thresholds",
        default=".gauntlet/thresholds.yml",
        help="path to thresholds YAML containing metrics.crap_ceiling.value",
    )
    parser.add_argument(
        "--ceiling",
        type=float,
        default=None,
        help="override CRAP ceiling (otherwise read from --thresholds)",
    )
    parser.add_argument(
        "--changed",
        default="",
        help="git ref; when set, scope to files changed since this ref",
    )
    parser.add_argument("--format", default="text", choices=("text", "json"))
    return parser.parse_args(argv)


def _report_to_json_object(report: Report) -> dict[str, object]:
    return {
        "ceiling": report.ceiling,
        "functions": [
            {
                "file": fn.file,
                "line": fn.line,
                "func": fn.func,
                "complexity": fn.complexity,
                "coverage": fn.coverage,
                "crap": fn.crap,
                "pass": fn.pass_,
            }
            for fn in report.functions
        ],
        "summary": {
            "total": report.summary.total,
            "failing": report.summary.failing,
            "max_crap": report.summary.max_crap,
        },
    }


def print_json_report(report: Report, stdout: TextIO) -> None:
    json.dump(_report_to_json_object(report), stdout, indent=2)
    stdout.write("\n")


def print_text_report(report: Report, stdout: TextIO) -> None:
    for fn in report.functions:
        status = "PASS" if fn.pass_ else "FAIL"
        note = "" if fn.matched else " (no coverage data)"
        stdout.write(
            f"{fn.file}:{fn.line}:{fn.func}\tcomplexity={fn.complexity}\t"
            f"coverage={fn.coverage:.1f}%\tcrap={fn.crap:.2f}\t{status}{note}\n"
        )
    stdout.write(
        f"\nceiling={report.ceiling:.2f} total={report.summary.total} "
        f"failing={report.summary.failing} max_crap={report.summary.max_crap:.2f}\n"
    )


def run(argv: Sequence[str], stdout: TextIO, stderr: TextIO) -> int:
    try:
        options = parse_options(argv, stderr)
        ceiling = options.ceiling
        if ceiling is None:
            ceiling = read_ceiling_from_thresholds(options.thresholds)

        functions = scan_dir(options.dir)
        if options.changed:
            changed_files = git_changed_files(options.dir, options.changed)
            if not changed_files:
                stderr.write(
                    f"crap4py: no files changed since {options.changed!r}; "
                    "nothing was scored\n"
                )
            functions = filter_by_changed_files(functions, changed_files)
        coverage = load_coverage_xml(options.coverage, options.dir, functions)
        report = evaluate(functions, coverage, ceiling)

        if options.format == "json":
            print_json_report(report, stdout)
        else:
            print_text_report(report, stdout)

        if report.summary.failing > 0:
            return 1
        return 0
    except _ParserExit as error:
        return error.status
    except (OSError, SyntaxError, ValueError, ET.ParseError) as error:
        stderr.write(f"crap4py: {error}\n")
        return 2


def main() -> int:
    return run(os.sys.argv[1:], os.sys.stdout, os.sys.stderr)


if __name__ == "__main__":
    raise SystemExit(main())
