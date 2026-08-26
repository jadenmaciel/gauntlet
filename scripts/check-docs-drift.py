#!/usr/bin/env python3
"""Fail CI when README, SKILL, or CONSUME disagree with adapters.yml."""

from __future__ import annotations

import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from lib.manifest import SCORER_SOURCES, check_flags_in_source, load_manifest

DOCS = ["README.md", "SKILL.md"]

CURRENT_TAG = "v0.2.0"
STALE_TAG_PATTERN = re.compile(r"@v0\.1\.0|--tag v0\.1\.0|branch v0\.1\.0|VERSION \?= v0\.1\.0")

CEILING_SCRAPE_PATTERN = re.compile(r"awk[^\n]*crap_ceiling")


def read(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


def check_manifest_commands(manifest: dict, failures: list[str]) -> None:
    for doc in DOCS:
        body = read(doc)
        for language, entry in manifest["languages"].items():
            for field in ("install", "coverage_command", "score"):
                command = entry[field]
                if command not in body:
                    failures.append(
                        f"{doc}: {language}.{field} from adapters.yml is not quoted verbatim.\n"
                        f"    expected: {command}"
                    )


def check_flags_exist(manifest: dict, failures: list[str]) -> None:
    for failure in check_flags_in_source(manifest):
        failures.append(failure)


def check_actions_exist(manifest: dict, failures: list[str]) -> None:
    for language, entry in manifest["languages"].items():
        reference = entry["action"]
        path_part, _, tag = reference.partition("@")
        relative = path_part.split("/", 2)[2]
        action = ROOT / relative / "action.yml"
        if not action.is_file():
            failures.append(f"adapters.yml: {language}.action points at missing {action.relative_to(ROOT)}")
        if tag != CURRENT_TAG:
            failures.append(f"adapters.yml: {language}.action pins {tag}, expected {CURRENT_TAG}")


def check_templates_exist(manifest: dict, failures: list[str]) -> None:
    for language in manifest["languages"]:
        fragment = ROOT / "templates" / language / "Makefile.fragment"
        if not fragment.is_file():
            failures.append(f"missing {fragment.relative_to(ROOT)} for language {language}")


def tracked_text_files() -> list[Path]:
    roots = [ROOT / "templates", ROOT / ".github"]
    files = [ROOT / doc for doc in DOCS] + [ROOT / "AGENTS.md", ROOT / "adapters.yml"]
    for directory in roots:
        files.extend(p for p in directory.rglob("*") if p.is_file())
    return files


def check_no_stale_pins(failures: list[str]) -> None:
    for path in tracked_text_files():
        try:
            body = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        if STALE_TAG_PATTERN.search(body):
            failures.append(f"{path.relative_to(ROOT)}: still pins v0.1.0, which has no adapters")


def check_no_ceiling_scraping(failures: list[str]) -> None:
    for path in tracked_text_files():
        try:
            body = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        for line in body.splitlines():
            if CEILING_SCRAPE_PATTERN.search(line) and "PLACEHOLDER" not in line:
                failures.append(
                    f"{path.relative_to(ROOT)}: scrapes crap_ceiling with awk; "
                    "pass --thresholds instead"
                )


def main() -> int:
    manifest = load_manifest()
    failures: list[str] = []

    check_manifest_commands(manifest, failures)
    check_flags_exist(manifest, failures)
    check_actions_exist(manifest, failures)
    check_templates_exist(manifest, failures)
    check_no_stale_pins(failures)
    check_no_ceiling_scraping(failures)

    if failures:
        print("docs drifted from adapters.yml:\n", file=sys.stderr)
        for failure in failures:
            print(f"  - {failure}", file=sys.stderr)
        print(
            f"\n{len(failures)} problem(s). adapters.yml is the source of truth; "
            "update the docs to match it.",
            file=sys.stderr,
        )
        return 1

    languages = ", ".join(manifest["languages"])
    print(f"docs agree with adapters.yml ({languages})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
