# CRAP contract

This file is the product contract for CRAP in gauntlet. Language adapters implement this file. They do not invent a second metric.

Audience: maintainers and internal consumers. The public, language-neutral docs are
[`README.md`](../README.md) for humans, [`SKILL.md`](../SKILL.md) for agents, and
[`adapters.yml`](../adapters.yml) for the commands.

## Formula

`internal/crap/crap.go` computes

`CRAP = CC² × (1 − cov)³ + CC`

`CC` is cyclomatic complexity. `cov` is coverage as a fraction from 0 to 1. `ComputeCRAP` takes coverage as a percent from 0 to 100 and divides by 100 inside the function.

The formula does not change.

## Join keys

`crap.Evaluate` joins complexity and coverage on `file:line:funcname`.

In `crap4go`, `file` in that key is the module-import path that `go tool cover -func` prints. Example: `github.com/jadenmaciel/gauntlet/internal/crap/crap.go`. The other four scorers key on repo-relative paths. That difference is why path matching is a path-segment suffix rather than string equality, and it is the only place the adapters legitimately differ.

A function with no coverage row gets coverage 0 and `matched` false. Evaluate does not drop it.

`--changed` keeps functions whose `file` matches a `git diff --name-only` path by exact match or by a path-segment suffix. `git diff` paths are relative to the repo root. Scanned files may be module-import paths.

## Command surface

All five scorers — `crap4go`, `crap4py`, `crap4rs`, `crap4ts`, `crap4php` — take the same flags:

```
<scorer> [--dir <path>] [--coverage <file>] [--thresholds <file>]
         [--ceiling <number>] [--changed <ref>] [--format text|json]
```

Aliases kept for pre-v0.2.0 callers: `crap4go --profile`, `crap4rs --coverage-json`,
`crap4rs --ceiling-file`.

## Report

`--format json` prints `ceiling`, `functions`, and `summary`.

Each function has `file`, `line`, `func`, `complexity`, `coverage`, `crap`, and `pass`.

`summary` has `total`, `failing`, and `max_crap`.

The key set is identical across the five scorers. The numeric rendering is not — Go prints a
whole ceiling as `30`, Python as `30.0` — so compare parsed JSON, never the raw bytes.

Exit codes: `0` pass, `1` at least one function over the ceiling, `2` usage or I/O error.
Exit 2 covers an unreadable `--coverage` path and an unresolvable ceiling; neither is allowed
to look like a pass.

## Thresholds

Ceilings live in `.gauntlet/thresholds.yml`.

```yaml
metrics:
  crap_ceiling:
    direction: max
    value: 30
```

`direction: max` means the number may only fall after the first baseline is recorded. `gauntlet ratchet` moves a ceiling down or a floor up.

Ceiling resolution, in order: `--ceiling` wins; otherwise the `--thresholds` file is read;
otherwise the scorer exits 2. No scorer carries a built-in default. Before v0.2.0 `crap4go`
defaulted to 8 and could not read the thresholds file at all, so it gated at 8 while the
committed file said 30.

## Ceiling policy E

Policy E, stated plainly: set the first `crap_ceiling.value` from a measured baseline on the
intended scope, then tighten it only with `gauntlet ratchet`.

Measure it with the scorer itself:

```bash
crap4go --dir . --coverage coverage.out --ceiling 100000 --format json
gauntlet init --crap-ceiling <summary.max_crap>
```

`gauntlet init` without `--crap-ceiling` writes 8 and labels it `PLACEHOLDER` in the generated
file. That is a refusal to guess, not a recommendation.

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

A Cursor Cloud VM and GitHub Actions run the same scorer command after installing that scorer and `gauntlet` at `v0.2.0`.

The environment fetches the base ref before `--changed`. A shallow clone without that ref is not done.

The job fails when any scored function exceeds `crap_ceiling.value`.

## Related documents

- Plan: [docs/plans/2026-08-20-uncle-bob-crap-cloud/overview.md](plans/2026-08-20-uncle-bob-crap-cloud/overview.md)
- Experiment citation: [docs/uncle-bob-negative-test-experiment.md](uncle-bob-negative-test-experiment.md)
- Consume recipe: [templates/CONSUME.md](../templates/CONSUME.md)
