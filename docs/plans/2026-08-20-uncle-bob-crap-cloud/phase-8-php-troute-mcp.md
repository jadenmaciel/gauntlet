# Phase 8. PHP path for troute-mcp

Back-link: [overview.md](overview.md)

## Goal

Make PHPUnit coverage real, then gate CRAP with a PHP adapter.

## Changes

- Replace Codecov TODO stub with clover/cobertura output in CI.
- Implement `crap4php` (complexity from PHP AST or phpmd/phploc-compatible source; coverage from clover).
- Golden vectors match Go `ComputeCRAP`.
- Wire `.gauntlet/thresholds.yml`, CI, Cursor install on `troute-mcp`.

## Data structures

- Same report JSON. Measured baseline ceiling.

## Verification

- Static. Coverage artifact non-empty in CI.
- Runtime. Intentional CRAP fail then fix. Control. GHA + Cloud if `environment.json` present.
