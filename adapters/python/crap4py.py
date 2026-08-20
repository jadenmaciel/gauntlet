#!/usr/bin/env python3
from __future__ import annotations

import argparse
import ast
import json
import os
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
        self.value += 1


def _is_wildcard_match_case(case: ast.match_case) -> bool:
    return isinstance(case.pattern, ast.MatchAs) and case.pattern.name is None


def function_complexity(node: ast.FunctionDef | ast.AsyncFunctionDef) -> int:
    counter = _ComplexityCounter()
    for statement in node.body:
        counter.visit(statement)
    return counter.value


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


def load_coverage_xml(path: str, root_dir: str) -> dict[str, CoverageEntry]:
    tree = ET.parse(path)
    root = tree.getroot()
    coverage: dict[str, CoverageEntry] = {}

    for class_node in root.findall(".//class"):
        filename = class_node.get("filename")
        if not filename:
            continue
        normalized_file = _normalize_file_path(filename, root_dir)
        for method_node in class_node.findall("./methods/method"):
            function_name = method_node.get("name")
            if not function_name:
                continue
            line_numbers: list[int] = []
            covered = 0
            total = 0
            for line_node in method_node.findall("./lines/line"):
                line_value = line_node.get("number")
                if not line_value:
                    continue
                line_numbers.append(int(line_value))
                total += 1
                if int(line_node.get("hits", "0")) > 0:
                    covered += 1
            if total == 0 or not line_numbers:
                continue
            fraction = covered / total
            key = build_key(normalized_file, min(line_numbers), function_name)
            coverage[key] = CoverageEntry(coverage=fraction)

    return coverage


def evaluate(
    functions: Iterable[FunctionComplexity], coverage: dict[str, CoverageEntry], ceiling: float
) -> Report:
    report_functions: list[FunctionReport] = []
    failing = 0
    max_crap = 0.0

    for fn in functions:
        entry = coverage.get(build_key(fn.file, fn.line, fn.func))
        matched = entry is not None
        coverage_fraction = entry.coverage if entry else 0.0
        score = compute_crap(fn.complexity, coverage_fraction * 100.0)
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
                coverage=coverage_fraction,
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
            f"coverage={fn.coverage * 100:.1f}%\tcrap={fn.crap:.2f}\t{status}{note}\n"
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
        coverage = load_coverage_xml(options.coverage, options.dir)
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
