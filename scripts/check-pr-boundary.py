#!/usr/bin/env python3
"""PR diff checks for architecture boundary violations."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


def changed_files(base_ref: str) -> list[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", f"{base_ref}...HEAD"],
        cwd=ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != 0:
        result = subprocess.run(
            ["git", "diff", "--name-only", base_ref],
            cwd=ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]


def main() -> int:
    base_ref = sys.argv[1] if len(sys.argv) > 1 else "origin/main"
    files = changed_files(base_ref)

    internal = [f for f in files if f.startswith("internal/")]
    adapters = [f for f in files if f.startswith("adapters/")]
    failures: list[str] = []

    if internal and adapters:
        failures.append(
            "PR touches both internal/ and adapters/ — split into atomic PRs "
            f"(internal: {len(internal)} files, adapters: {len(adapters)} files)"
        )

    new_go = [
        f
        for f in files
        if f.endswith(".go")
        and not f.endswith("_test.go")
        and f.startswith(("internal/", "adapters/", "cmd/"))
    ]
    for path in new_go:
        test_path = Path(path).with_name(f"{Path(path).stem}_test.go")
        if not (ROOT / test_path).is_file() and not any(
            t.endswith("_test.go") and t.startswith(str(Path(path).parent)) for t in files
        ):
            failures.append(f"{path}: new Go source without a colocated _test.go in this PR")

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1

    print("PR boundary OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
