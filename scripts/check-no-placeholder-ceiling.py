#!/usr/bin/env python3
"""Fail CI when the root thresholds file is a placeholder, not a measured baseline."""

from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
THRESHOLDS = ROOT / ".gauntlet" / "thresholds.yml"


def main() -> int:
    if not THRESHOLDS.is_file():
        print(f"missing {THRESHOLDS.relative_to(ROOT)}", file=sys.stderr)
        return 1

    body = THRESHOLDS.read_text(encoding="utf-8")
    if "PLACEHOLDER" in body:
        print(
            f"{THRESHOLDS.relative_to(ROOT)}: contains PLACEHOLDER; "
            "measure with --ceiling 100000 and gauntlet init --crap-ceiling",
            file=sys.stderr,
        )
        return 1

    if "Baseline measured" not in body and "crap_ceiling" in body:
        for line in body.splitlines():
            stripped = line.strip()
            if stripped.startswith("value:") and stripped.split(":", 1)[1].strip() in {"8", "8.0"}:
                print(
                    f"{THRESHOLDS.relative_to(ROOT)}: value 8 without measured baseline comment",
                    file=sys.stderr,
                )
                return 1

    print(f"{THRESHOLDS.relative_to(ROOT)} is not a placeholder")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
