# Phase 7. TypeScript path for purely-expo

Back-link: [overview.md](overview.md)

## Goal

Make coverage real, then gate CRAP with a TS adapter. Do not claim CRAP while Codecov is a stub.

## Changes

- Enable coverage on the packages you care about (`@purely/analysis` and any app packages in scope).
- Implement `crap4ts` (complexity from a TS/ESLint complexity source or similar; coverage from Istanbul/V8).
- Golden vectors match Go `ComputeCRAP`.
- Wire `.gauntlet/thresholds.yml`, CI, Cursor install.

## Data structures

- Same report JSON. Thresholds with measured baseline ceiling.

## Verification

- Static. Coverage job produces a real profile. Adapter unit tests green.
- Runtime. Intentional CRAP fail then fix on a PR. Control. GHA.
