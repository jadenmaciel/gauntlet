#!/usr/bin/env python3
"""Fail CI when Go import boundaries are violated.

Rules (see docs/ARCHITECTURE.md):
  - internal/* must not import adapters/*
  - cmd/* is the only tree that may import internal/*
  - adapters are polyglot; this script checks Go only
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

IMPORT_RE = re.compile(r'"([^"]+)"')


def go_files_under(prefix: str) -> list[Path]:
    base = ROOT / prefix
    if not base.is_dir():
        return []
    return sorted(base.rglob("*.go"))


def parse_imports(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    imports: list[str] = []
    in_block = False
    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("import ("):
            in_block = True
            continue
        if in_block:
            if stripped == ")":
                in_block = False
                continue
            match = IMPORT_RE.search(stripped)
            if match:
                imports.append(match.group(1))
            continue
        if stripped.startswith("import "):
            match = IMPORT_RE.search(stripped)
            if match:
                imports.append(match.group(1))
    return imports


def rel(path: Path) -> str:
    return path.relative_to(ROOT).as_posix()


def main() -> int:
    failures: list[str] = []

    for path in go_files_under("internal"):
        for imp in parse_imports(path):
            if "/adapters/" in imp or imp.endswith("/adapters"):
                failures.append(f"{rel(path)}: internal must not import adapter {imp}")

    adapter_go = go_files_under("adapters")
    for path in adapter_go:
        for imp in parse_imports(path):
            if "/internal/" in imp:
                failures.append(f"{rel(path)}: adapter must not import internal {imp}")

    for path in go_files_under("cmd"):
        for imp in parse_imports(path):
            if "/adapters/" in imp:
                failures.append(f"{rel(path)}: cmd must not import adapters directly {imp}")

    for path in ROOT.rglob("*.go"):
        rel_path = rel(path)
        if rel_path.startswith(("cmd/", "internal/")):
            continue
        for imp in parse_imports(path):
            if "github.com/jadenmaciel/gauntlet/internal" in imp:
                failures.append(f"{rel_path}: only cmd/* may import internal/*, not {imp}")

    if failures:
        print("import graph violations:\n", file=sys.stderr)
        for failure in failures:
            print(f"  - {failure}", file=sys.stderr)
        return 1

    print("Go import graph OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
