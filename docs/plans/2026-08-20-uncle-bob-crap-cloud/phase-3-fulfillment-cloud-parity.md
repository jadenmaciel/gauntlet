# Phase 3. Fulfillment cloud parity

Back-link: [overview.md](overview.md)

## Goal

Prove Cursor Cloud for `troute-fulfillment` can run `make crap` with the same pins and ref as CI.

## Changes

- Extend `.cursor/cloud-env.sh install` (if needed) so `bin/crap4go` and `bin/gauntlet` exist after install, matching Makefile pins.
- Ensure cloud boot fetches `origin/develop` (already partially true) so `--changed origin/develop` is meaningful.
- Document the cloud check path in `docs/dev/cursor-cloud.md` (`make crap` or `cloud-env.sh check` including crap).
- Optionally call the new composite Action from CI once phase 2 lands (or keep Makefile-only if Action is thin wrapper).

## Data structures

- No threshold change unless a measured baseline says otherwise.

## Verification

- Static. `make check` still green locally.
- Runtime. Start a Cursor Cloud agent on fulfillment. Run `make crap`. Confirm exit 0 and that the binary is the pinned VM install, not a host path. Control. Cloud agent session (prove-it-works on the real surface).
