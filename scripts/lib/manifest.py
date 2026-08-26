"""Shared adapters.yml helpers for contract check scripts."""

from __future__ import annotations

from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[2]

SCORER_SOURCES = {
    "go": "cmd/crap4go/main.go",
    "python": "adapters/python/crap4py.py",
    "rust": "adapters/rust/crap4rs/src/cli.rs",
    "typescript": "adapters/typescript/crap4ts/src/cli.ts",
    "php": "adapters/php/crap4php/src/OptionParser.php",
}


def load_manifest() -> dict:
    return yaml.safe_load((ROOT / "adapters.yml").read_text(encoding="utf-8"))


def read_source(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


def check_flags_in_source(manifest: dict) -> list[str]:
    failures: list[str] = []
    flags = list(manifest["cli"]["flags"])
    for language, source in SCORER_SOURCES.items():
        body = read_source(source)
        for flag in flags:
            if f"--{flag}" not in body and f'"{flag}"' not in body and f"-{flag}" not in body:
                failures.append(f"{source}: {language} scorer never mentions --{flag}")
    return failures
