#!/usr/bin/env python3
"""Run each scorer against a fixture and assert CLI JSON contract from adapters.yml."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from lib.manifest import check_flags_in_source, load_manifest


def check_json_shape(payload: dict, manifest: dict) -> list[str]:
    failures: list[str] = []
    root_keys = set(manifest["cli"]["json_keys"]["root"])
    fn_keys = set(manifest["cli"]["json_keys"]["function"])
    summary_keys = set(manifest["cli"]["json_keys"]["summary"])

    if set(payload.keys()) != root_keys:
        failures.append(f"root keys {sorted(payload.keys())} != {sorted(root_keys)}")

    if set(payload.get("summary", {}).keys()) != summary_keys:
        failures.append(
            f"summary keys {sorted(payload.get('summary', {}).keys())} != {sorted(summary_keys)}"
        )

    for function in payload.get("functions", []):
        if set(function.keys()) != fn_keys:
            failures.append(
                f"function keys {sorted(function.keys())} != {sorted(fn_keys)}"
            )
            break
    return failures


def run(cmd: list[str], *, cwd: Path | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        cmd,
        cwd=cwd or ROOT,
        text=True,
        capture_output=True,
        check=False,
    )


def check_go(manifest: dict) -> list[str]:
    failures: list[str] = []
    build = run(["go", "build", "-o", "/tmp/crap4go-contract", "./cmd/crap4go"])
    if build.returncode != 0:
        return [f"go build crap4go: {build.stderr.strip()}"]

    cov = run(["go", "test", "-coverprofile=/tmp/gauntlet-contract-cov.out", "./..."])
    if cov.returncode != 0:
        failures.append(f"go test coverage: {cov.stderr.strip()}")
        return failures

    score = run(
        [
            "/tmp/crap4go-contract",
            "--dir",
            ".",
            "--coverage",
            "/tmp/gauntlet-contract-cov.out",
            "--thresholds",
            ".gauntlet/thresholds.yml",
            "--format",
            "json",
        ]
    )
    if score.returncode not in (0, 1):
        failures.append(f"crap4go scoring exited {score.returncode}: {score.stderr.strip()}")
        return failures

    try:
        payload = json.loads(score.stdout)
    except json.JSONDecodeError as err:
        return [f"crap4go JSON parse: {err}"]

    failures.extend(check_json_shape(payload, manifest))
    return failures


def check_python(manifest: dict) -> list[str]:
    failures: list[str] = []
    fixture = ROOT / "adapters/python/tests/data/demo_hotspot"
    cli = ROOT / "adapters/python/crap4py.py"

    score = run(
        [
            sys.executable,
            str(cli),
            "--dir",
            str(fixture),
            "--coverage",
            str(fixture / "coverage-low.xml"),
            "--thresholds",
            str(fixture / ".gauntlet/thresholds.yml"),
            "--format",
            "json",
        ]
    )
    if score.returncode not in (0, 1):
        failures.append(f"crap4py scoring exited {score.returncode}: {score.stderr.strip()}")
        return failures

    try:
        payload = json.loads(score.stdout)
    except json.JSONDecodeError as err:
        return [f"crap4py JSON parse: {err}"]

    failures.extend(check_json_shape(payload, manifest))
    return failures


def check_rust(manifest: dict) -> list[str]:
    failures: list[str] = []
    manifest_path = ROOT / "adapters/rust/crap4rs/Cargo.toml"

    build = run(["cargo", "build", "--manifest-path", str(manifest_path), "--quiet"])
    if build.returncode != 0:
        return [f"cargo build crap4rs: {build.stderr.strip()}"]

    unit = run(
        [
            "cargo",
            "test",
            "--manifest-path",
            str(manifest_path),
            "emits_json_report_shape",
            "--quiet",
        ]
    )
    if unit.returncode != 0:
        failures.append(f"crap4rs JSON shape test: {unit.stderr.strip()}")
    return failures


def check_typescript(manifest: dict) -> list[str]:
    failures: list[str] = []
    pkg = ROOT / "adapters/typescript/crap4ts"
    fixture = pkg / "demo/project"
    coverage = pkg / "demo/coverage-low.json"

    build = run(["npm", "run", "build", "--silent"], cwd=pkg)
    if build.returncode != 0:
        build = run(["pnpm", "run", "build"], cwd=pkg)
    if build.returncode != 0:
        return [f"build crap4ts: {build.stderr.strip()}"]

    cli = pkg / "dist/cli.js"

    score = run(
        [
            "node",
            str(cli),
            "--dir",
            str(fixture),
            "--coverage",
            str(coverage),
            "--thresholds",
            str(fixture / ".gauntlet/thresholds.yml"),
            "--format",
            "json",
        ]
    )
    if score.returncode not in (0, 1):
        failures.append(f"crap4ts scoring exited {score.returncode}: {score.stderr.strip()}")
        return failures

    try:
        payload = json.loads(score.stdout)
    except json.JSONDecodeError as err:
        return [f"crap4ts JSON parse: {err}"]

    failures.extend(check_json_shape(payload, manifest))
    return failures


def check_php(manifest: dict) -> list[str]:
    failures: list[str] = []
    fixture = ROOT / "adapters/php/crap4php/tests/fixtures"
    pkg = ROOT / "adapters/php/crap4php"
    bin_path = pkg / "bin/crap4php"

    score = run(
        [
            "php",
            str(bin_path),
            "--dir",
            str(fixture),
            "--coverage",
            str(fixture / "clover-high.xml"),
            "--ceiling",
            "30",
            "--format",
            "json",
        ]
    )
    if score.returncode not in (0, 1):
        unit = run(
            ["composer", "test", "--filter", "testJsonFormatMatchesContractShape"],
            cwd=pkg,
        )
        if unit.returncode != 0:
            detail = score.stderr.strip() or unit.stderr.strip() or "php contract check failed"
            failures.append(detail)
        return failures

    try:
        payload = json.loads(score.stdout)
    except json.JSONDecodeError as err:
        return [f"crap4php JSON parse: {err}"]

    failures.extend(check_json_shape(payload, manifest))
    return failures


def main() -> int:
    manifest = load_manifest()
    failures: list[str] = check_flags_in_source(manifest)

    for name, checker in (
        ("go", check_go),
        ("python", check_python),
        ("rust", check_rust),
        ("typescript", check_typescript),
        ("php", check_php),
    ):
        lang_failures = checker(manifest)
        for failure in lang_failures:
            failures.append(f"{name}: {failure}")

    if failures:
        print("CLI contract failures:\n", file=sys.stderr)
        for failure in failures:
            print(f"  - {failure}", file=sys.stderr)
        return 1

    print("CLI contract OK (all scorers)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
