# Phase 5. Python `crap4py` for clark-agency

Back-link: [overview.md](overview.md)

## Goal

Add an honest Python CRAP scorer that reuses `internal/crap` math (or a shared golden-vector table) and wire it into `clark-agency`.

## Changes

- Implement `crap4py` in `gauntlet` (complexity from AST or radon; coverage from `coverage.py` XML / JSON). Join by stable file:line:func keys documented in phase 1.
- Add `templates/python/` Makefile or make-target fragment.
- On `clark-agency`. Add `crap_ceiling` to `.gauntlet/thresholds.yml`, `make crap` into `make check`, CI job, and Cursor `environment.json` install for the tool.
- Golden tests. Same numeric vectors as `ComputeCRAP` in Go.

## Data structures

- Shared report JSON shape with `crap4go` so Actions can stay format-stable.
- Thresholds gain `crap_ceiling` beside existing `coverage_min`.

## Verification

- Static. Golden CRAP vectors match Go. clark-agency lint/tests green.
- Runtime. `make crap` fails a planted hot function, passes after split/cover. Cloud install runs the same binary/module. Control. `control-cli` / `make check` in clark-agency; one Cloud agent smoke.
