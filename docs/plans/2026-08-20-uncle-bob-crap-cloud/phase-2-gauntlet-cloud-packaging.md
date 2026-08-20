# Phase 2. Gauntlet cloud packaging

Back-link: [overview.md](overview.md)

## Goal

Make `gauntlet` installable the same way in CI and on Cursor Cloud VMs without a laptop `GOPATH`. Include a **future-repo** consume recipe so new repos copy one path.

## Changes

- Add a real README (install pins at `v0.1.0`, Makefile fragment usage, thresholds, consumer list).
- Add a composite GitHub Action under `.github/actions/crap-go` that installs `crap4go`, expects a coverage profile path, reads ceiling from `.gauntlet/thresholds.yml`, and fails the job on failing functions.
- Add `templates/cursor/cloud-install-go.sh` consumers paste into `.cursor/cloud-env.sh` / `environment.json` `install`.
- Add `templates/CONSUME.md` for future repos (thresholds file, language adapter pick, Action call, Cloud install).
- Keep fragment defaults from floating `latest` to the pinned tag.

## Data structures

- Action inputs. `coverage-profile`, `ceiling-file` (default `.gauntlet/thresholds.yml`), `changed-ref`, `working-directory`.
- No new CRAP formula fields.

## Verification

- Static. Action YAML validates. `go test ./...` still green.
- Runtime. Dry-run or branch workflow. Plant a high-CRAP function, see Action fail, remove it, see pass. Control. GHA on a branch or `act` if available.
