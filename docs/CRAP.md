# CRAP contract

This file is the product contract for CRAP in gauntlet. Language adapters implement this file. They do not invent a second metric.

## Formula

`internal/crap/crap.go` computes

`CRAP = CC² × (1 − cov)³ + CC`

`CC` is cyclomatic complexity. `cov` is coverage as a fraction from 0 to 1. `ComputeCRAP` takes coverage as a percent from 0 to 100 and divides by 100 inside the function.

The formula does not change.

## Join keys

`crap.Evaluate` joins complexity and coverage on `file:line:funcname`.

`file` in that key is the module-import path that `go tool cover -func` prints. Example: `github.com/jadenmaciel/gauntlet/internal/crap/crap.go`.

A function with no coverage row gets coverage 0 and `matched` false. Evaluate does not drop it.

`--changed` keeps functions whose `file` matches a `git diff --name-only` path by exact match or by a path-segment suffix. `git diff` paths are relative to the repo root. Scanned files may be module-import paths.

## Report

`crap4go --format json` prints `ceiling`, `functions`, and `summary`.

Each function has `file`, `line`, `func`, `complexity`, `coverage`, `crap`, and `pass`.

`summary` has `total`, `failing`, and `max_crap`.

`crap4go` exits 1 when `summary.failing` is greater than 0.

## Thresholds

Ceilings live in `.gauntlet/thresholds.yml`.

```yaml
metrics:
  crap_ceiling:
    direction: max
    value: 30
```

`direction: max` means the number may only fall after the first baseline is recorded. `gauntlet ratchet` moves a ceiling down or a floor up.

## Ceiling policy E

Set the first `crap_ceiling.value` from a measured baseline on the intended scope. After that, tighten it only with `gauntlet ratchet`.

`troute-fulfillment` measured and ships at 30 with `--changed origin/develop`. That 30 is the Go precedent. It is not a new formula. It is not a day-one cap of 4.

Product CI does not force ceiling 4 on a brownfield tree. The experiment used CRAP below 4 as a research knob.

## Do not expand

Product CI is the CRAP constraint only. The score stays `ComputeCRAP`. Mutation is not a pull-request gate in this wave. Hunt the Wumpus is not a product test. SCRAP and hybrid scores stay out.

## Consumers

- `troute-fulfillment`. Go. Live gate. Ceiling 30.
- `troute-mcp`. PHP. Later adapter.
- `troute-comms`. Rust. Later adapter.
- `purely-expo`. TypeScript. Later adapter.
- `clark-agency`. Python. Later adapter.
- Future repos. Follow `templates/CONSUME.md`.

## Cloud done when

A Cursor Cloud VM and GitHub Actions run the same `crap4go` command after installing `crap4go` and `gauntlet` at `v0.1.0`.

The environment fetches the base ref before `--changed`. A shallow clone without that ref is not done.

The job fails when any scored function exceeds `crap_ceiling.value`.

## Related documents

- Plan: [docs/plans/2026-08-20-uncle-bob-crap-cloud/overview.md](plans/2026-08-20-uncle-bob-crap-cloud/overview.md)
- Experiment citation: [docs/uncle-bob-negative-test-experiment.md](uncle-bob-negative-test-experiment.md)
- Consume recipe: [templates/CONSUME.md](../templates/CONSUME.md)
