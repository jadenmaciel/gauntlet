# Phase 9 — Swift spike (Purely)

Back-link: [overview.md](overview.md)

## Goal

Decide whether an honest Swift CRAP gate is feasible on Xcode Cloud, or stop with a documented deferral.

## Changes

- Spike only: xccov function coverage + a Swift complexity source.
- If join keys cannot be trusted, write a deferral note under gauntlet docs and leave Purely on coverage hatch only.
- Do not ship a fake gate.

## Data structures

- Spike notes. Adapter only if the join is proven.

## Verification

- Runtime: one local `xcodebuild` + coverage export proving function-level coverage exists. Complexity join either works or the phase ends as “deferred.”
