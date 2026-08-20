# Phase 5 — Pilot troute-shipping

Back-link: [overview.md](overview.md)

## Goal

`troute-shipping` gets the same Python CRAP gate and a real Cursor `environment.json` (missing today).

## Changes

- Introduce `.gauntlet/thresholds.yml` and a check entrypoint (Makefile or extend `scripts/repo-gate.sh`).
- Hook CI `.github/workflows/ci.yml` after existing pytest-cov.
- Add `.cursor/environment.json` + install that pulls the pinned adapter and fetches the base branch.

## Data structures

- Same thresholds shape as clark-agency.

## Verification

- Static: CI job runs CRAP on PRs.
- Runtime: Cloud Agent or local install script runs the gate once; plant-and-fail proof on a throwaway branch.
