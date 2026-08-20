# Phase 1. Baseline and contract

Back-link: [overview.md](overview.md)

## Goal

Write down the non-negotiable CRAP contract so every later phase implements the same thing.

## Changes

- Add `docs/CRAP.md` in `gauntlet` with the formula, join-key rules, ceiling policy, and explicit “do not invent variants” line.
- Point at Uncle Bob’s experiment for history. State that product adoption is the constraint gate, not the HTW grid.
- Record fulfillment as the reference consumer (ceiling `30`, `--changed origin/develop`).

## Data structures

- Thresholds stay the existing nested YAML shape under `.gauntlet/thresholds.yml` (`crap_ceiling.direction: max`, `value: <number>`).
- Report schema stays `crap4go` JSON (`ceiling`, `functions[]`, `summary`).

## Verification

- Static. Doc exists. Formula matches `internal/crap/crap.go`.
- Runtime. N/A (docs only). Control surface. none.
