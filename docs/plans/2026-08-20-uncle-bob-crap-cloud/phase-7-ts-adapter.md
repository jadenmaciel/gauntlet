# Phase 7 — TypeScript adapter

Back-link: [overview.md](overview.md)

## Goal

Same formula for TS/JS once coverage exists (purely-expo wave 2; shipping frontend optional).

## Changes

- Adapter joins cyclomatic complexity (eslint complexity or typhonjs-escomplex class of tool) with istanbul/v8 coverage.
- Template + Action input for Node repos.
- Pilot on one small package before monorepo-wide.

## Data structures

- Same report schema as Go/Python.

## Verification

- Golden vector parity with `ComputeCRAP`.
- Plant-and-fail on the pilot package.
