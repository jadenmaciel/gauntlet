# Phase 6 — Coverage prerequisites

Back-link: [overview.md](overview.md)

## Goal

Unblock CRAP where the analyzer cannot run because coverage is off or stubbed.

## Changes

- `epayment`: enable PHPUnit coverage driver in CI (pcov/xdebug), produce clover/cobertura artifact.
- `troute-mcp`: real coverage config (not Codecov TODO stub).
- `purely-expo`: coverage reporter on packages that will be gated (at least `@purely/analysis`).

Stop after green coverage floors. Do not invent CRAP numbers without complexity join.

## Data structures

- Coverage artifacts only (clover/lcov/cobertura). No CRAP types yet.

## Verification

- Static: CI publishes a coverage report artifact.
- Runtime: coverage percentage visible in CI logs; floor optional but recommended.
