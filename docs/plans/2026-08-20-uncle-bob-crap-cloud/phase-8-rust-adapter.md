# Phase 8 — Rust adapter (troute-comms)

Back-link: [overview.md](overview.md)

## Goal

Gate `troute-comms` with CRAP using llvm-cov plus a Rust complexity source.

## Changes

- Choose a maintained complexity measure that maps to per-function CC (document the mapping).
- Join with `cargo llvm-cov` output.
- Wire CI job alongside the existing 88% line floor.
- Confirm `.cursor/environment.json` installs the tool.

## Data structures

- Same report schema.

## Verification

- Golden vectors + plant-and-fail on a throwaway function in troute-comms.
