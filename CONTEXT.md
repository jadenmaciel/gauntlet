# gauntlet

Multi-language CRAP gate: five interchangeable scorers, one manifest, ratcheting ceilings. Repo is the product agents and consumer CI install from.

## Language

**Adapter**:
A language-specific scorer package under `adapters/<lang>/` plus its composite GitHub Action and Makefile fragment.
_Avoid_: plugin, driver, wrapper (without naming the language)

**Scorer**:
The CLI binary that reads coverage and source, emits CRAP scores, and exits non-zero when a function exceeds the ceiling (`crap4go`, `crap4py`, `crap4rs`, `crap4ts`, `crap4php`).
_Avoid_: linter, analyzer (generic)

**Thresholds file**:
The durable YAML at `.gauntlet/thresholds.yml` holding `metrics.crap_ceiling.value` and other ratcheting metrics.
_Avoid_: config, settings file

**Composite action**:
A reusable GitHub Action under `.github/actions/crap-<lang>/` that installs a pinned scorer and runs it in consumer CI.
_Avoid_: workflow step (inline), reusable workflow

**Manifest**:
The machine-readable contract at `adapters.yml` — install commands, coverage format, CLI flags, JSON shape, exit codes.
_Avoid_: README, SKILL (as source of truth)

**Ratchet**:
The monotonic update that may only lower a ceiling or raise a floor; performed by `gauntlet ratchet`, never by hand-editing the thresholds file.
_Avoid_: threshold update, ceiling bump

## Relationships

- The **Manifest** defines how every **Scorer** must behave.
- Each **Adapter** implements one **Scorer** and publishes one **Composite action**.
- Consumer repos store a **Thresholds file** and invoke a **Composite action** in CI.
- Only `gauntlet init` and `gauntlet ratchet` write the **Thresholds file** ceiling.
