# Phase 4. Go consumer pilot

Back-link: [overview.md](overview.md)

## Goal

Prove the packaging works on a second Go tree so fulfillment is not a one-off snowflake.

## Changes

- Prefer gating `gauntlet` itself (dogfood) with `.gauntlet/thresholds.yml` + `make crap` + Action.
- If another main Go product repo appears, consume the fragment + Action instead of inventing a third pattern.
- Measure initial max CRAP on `--changed` scope. Set ceiling at or slightly above measured max, then ratchet only downward.

## Data structures

- Same thresholds schema as fulfillment.

## Verification

- Static. Unit tests + `make crap` on the pilot repo.
- Runtime. PR CI red on intentional violation, green on fix. Control. GitHub check run.
