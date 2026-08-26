#!/usr/bin/env python3
"""Collect trust-system PR signals for the pr-review workflow comment."""

from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


def run(cmd: list[str], *, check: bool = False) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, cwd=ROOT, text=True, capture_output=True, check=check)


def write_github_output(name: str, value: str) -> None:
    path = os.environ.get("GITHUB_OUTPUT")
    if not path:
        print(f"{name}={value}")
        return
    with open(path, "a", encoding="utf-8") as handle:
        handle.write(f"{name}<<EOF\n{value}\nEOF\n")


def main() -> int:
    base_ref = sys.argv[1] if len(sys.argv) > 1 else "origin/main"
    findings: list[str] = []

    boundary = run(["python3", "scripts/check-pr-boundary.py", base_ref])
    if boundary.returncode != 0:
        detail = boundary.stderr.strip() or boundary.stdout.strip() or "see CI logs"
        findings.append(
            f"- **Boundary:** PR touches both `internal/` and `adapters/`, "
            f"or adds Go source without tests. ({detail})"
        )

    drift = run(["python3", "scripts/check-docs-drift.py"])
    if drift.returncode != 0:
        findings.append("- **Docs drift:** README or SKILL diverged from `adapters.yml`.")

    tests = run(["go", "test", "-coverprofile=coverage.out", "./..."])
    if tests.returncode != 0:
        findings.append("- **Go tests:** failed before CRAP could run.")
    else:
        install = run(["go", "install", "./cmd/crap4go"])
        if install.returncode != 0:
            findings.append("- **CRAP check failed:** could not install crap4go.")
        else:
            score = run(
                [
                    "crap4go",
                    "--dir",
                    ".",
                    "--coverage",
                    "coverage.out",
                    "--thresholds",
                    ".gauntlet/thresholds.yml",
                    "--changed",
                    base_ref,
                    "--format",
                    "json",
                ]
            )
            if score.returncode not in (0, 1):
                err = score.stderr.strip() or "unknown scorer error"
                findings.append(f"- **CRAP check failed:** {err}")
            elif score.returncode == 1:
                try:
                    payload = json.loads(score.stdout)
                except json.JSONDecodeError:
                    findings.append("- **CRAP regression:** scorer returned invalid JSON.")
                else:
                    summary = payload.get("summary", {})
                    findings.append(
                        "- **CRAP regression:** "
                        f"{summary.get('failing', '?')} changed function(s) over ceiling "
                        f"(max CRAP {summary.get('max_crap', '?')})."
                    )

    if findings:
        body = "### Gauntlet trust review\n\n" + "\n".join(findings)
        write_github_output("status", "fail")
    else:
        body = "### Gauntlet trust review\n\nAll automated checks passed for this PR shape."
        write_github_output("status", "pass")

    write_github_output("body", body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
