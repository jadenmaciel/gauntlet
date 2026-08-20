# Phase 6. Rust adapter for troute-comms

Back-link: [overview.md](overview.md)

## Goal

Add an honest Rust CRAP scorer and wire it into `troute-comms` after `crap4py` proves multi-lang.

## Changes

- Implement `crap4rs` (or `gauntlet` subcommand) using cyclomatic complexity from a Rust AST/tool and `cargo llvm-cov` LCOV/JSON.
- Golden vectors must match `internal/crap.ComputeCRAP`.
- Add `templates/rust/` consume fragment.
- On `troute-comms`. Add `.gauntlet/thresholds.yml` with measured `crap_ceiling`, CI step, Cursor `environment.json` install.

## Data structures

- Same report JSON shape as `crap4go`.
- Thresholds. `crap_ceiling` + keep existing line-coverage floor (88%).

## Verification

- Static. Golden vectors. `cargo test` / workspace checks green.
- Runtime. Plant high-CRAP fn → fail; fix → pass. Cloud install smoke. Control. GHA + optional Cloud agent.
