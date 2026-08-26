# Architecture

Agents arrive with a narrow prompt and a few nearby files. They copy the nearest pattern, edit the open file, and take the shortest path that compiles. Soft reminders lose. Hard checks plus a verification loop win.

This document is the human/agent contract for working **on** gauntlet. Commands live in [`adapters.yml`](../adapters.yml). Adoption for other repos lives in [`SKILL.md`](../SKILL.md).

## Contributor context model

You will usually see one adapter folder, one CI job, or one scorer CLI — not the whole tree. Before inventing anything:

1. Read [`adapters.yml`](../adapters.yml) for the exact flags, JSON keys, and exit codes.
2. Copy the closest existing adapter; do not invent a sixth CLI dialect.
3. Run the language tests plus the repo contract checks under `scripts/`.
4. Produce evidence: scorer JSON output and exit code on a fixture, not "tests pass" alone.

## Five rules

### 1. Conventional path is the cheap path

| Change | Conventional path |
|---|---|
| New language | `adapters/<lang>/` + `adapters.yml` block + `.github/actions/crap-<lang>/` + CI job + template fragment |
| New CLI flag | All five scorers + `adapters.yml` + docs-drift green |
| New public skill | Not in this repo — consumer repos point at [`SKILL.md`](../SKILL.md) |

Ban "just add a flag to `crap4go`." If one scorer gets a flag, all five do, or CI fails.

### 2. Forbidden dependencies fail mechanically

| From | Must not import |
|---|---|
| `internal/crap`, `internal/thresholds` | `adapters/*` |
| Each language adapter | Any other adapter |
| Wiring | Only `cmd/*` imports `internal/*` |

CI enforces the Go import graph via `scripts/check-import-graph.py`. Polyglot adapters stay isolated by directory, not shared libraries.

### 3. One obvious writer

| Durable value | Only writer |
|---|---|
| `adapters.yml` | Human + docs-drift (prose must match) |
| `.gauntlet/thresholds.yml` ceiling | `gauntlet ratchet` / `gauntlet init` |
| `metrics.crap_ceiling.value` in consumer repos | Same — agents never hand-edit ceilings |

### 4. Isolated files, not shared-root branches

Language quirks live under `adapters/<lang>/`. Do not grow `internal/crap` into a dumping ground for per-language exceptions. Shared logic belongs in the formula and filter, not adapter-specific branches.

### 5. Exceptions are architecture PRs

Breaking an invariant (cross-adapter import, invented ceiling, comment-only "fix") requires an ADR under `docs/adr/` and an explicit CI allowlist entry. The allowlist starts empty.

## Public nouns

See [`CONTEXT.md`](../CONTEXT.md) for Adapter, Scorer, Thresholds file, Composite action, Manifest, Ratchet.

## Verification loop

| Change type | Evidence before "done" |
|---|---|
| Scorer / CLI flag | Language tests **and** installed binary on a fixture; JSON matches `adapters.yml` `json_keys` |
| New adapter | Golden vector + composite action dry-run |
| Docs | `scripts/check-docs-drift.py` green |
| Working on this repo | `go test ./...` + contract scripts + self-gate CRAP on changed Go functions |

## Layout

| path | role |
|---|---|
| `internal/crap/` | Formula, complexity, coverage join, changed-file filter |
| `internal/thresholds/` | Ratchet file load/save/verify |
| `cmd/crap4go/`, `cmd/gauntlet/` | Go scorer and ratchet CLI |
| `adapters/{python,rust,typescript,php}/` | Other four scorers |
| `adapters.yml` | Manifest |
| `.github/actions/crap-*/` | Composite actions |
| `scripts/` | Mechanical contract checks |
