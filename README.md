# gauntlet

`crap4go` scores Go functions with CRAP. `gauntlet` stores floors and ceilings that only tighten.

The product contract is [docs/CRAP.md](docs/CRAP.md). The rollout plan is [docs/plans/2026-08-20-uncle-bob-crap-cloud/overview.md](docs/plans/2026-08-20-uncle-bob-crap-cloud/overview.md). To add this gate to another repo, follow [templates/CONSUME.md](templates/CONSUME.md).

## Install

Pin both commands at `v0.1.0`.

```bash
go install github.com/jadenmaciel/gauntlet/cmd/crap4go@v0.1.0
go install github.com/jadenmaciel/gauntlet/cmd/gauntlet@v0.1.0
```

On Cursor Cloud, run [templates/cursor/cloud-install-go.sh](templates/cursor/cloud-install-go.sh). It installs both binaries into `$HOME/.local/bin`, or into `GAUNTLET_BIN_DIR` if you set that.

## Score Go with crap4go

```bash
go test -coverprofile=coverage.out ./...
crap4go --dir . --profile coverage.out --ceiling 30 --changed origin/main
```

`--changed` is optional. Empty means the whole tree. Fetch the base ref first on a shallow clone.

## Makefile fragment

Copy [templates/go/Makefile.fragment](templates/go/Makefile.fragment) into the repo Makefile. `CRAP4GO_VERSION` and `GAUNTLET_VERSION` default to `v0.1.0`. The fragment reads `metrics.crap_ceiling.value` from `.gauntlet/thresholds.yml` with awk.

## Thresholds

```yaml
metrics:
  crap_ceiling:
    direction: max
    value: 30
```

Set the first value from a measured baseline (policy E). After that:

```bash
gauntlet ratchet --metric crap_ceiling --value <new-lower-number>
```

`troute-fulfillment` ships at 30. That is the Go precedent, not a required copy.

## GitHub Action

```yaml
- uses: jadenmaciel/gauntlet/.github/actions/crap-go@v0.1.0
  with:
    profile: coverage.out
    ceiling-file: .gauntlet/thresholds.yml
    changed-ref: origin/main
```

The job needs Go on `PATH` before this step.
