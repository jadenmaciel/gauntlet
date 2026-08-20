# Uncle Bob CRAP across main repos (cloud-native)

**Status.** Decisions locked 2026-08-20. Implementation started on gauntlet phases 1–2.

## Locked decisions

| Decision | Lock |
|---|---|
| Consumers | `troute-fulfillment`, `troute-mcp`, `troute-comms`, `purely-expo`, `clark-agency`, plus a **future-repos** consume recipe |
| Ceiling (**E**) | Ratchet from measured baseline. Floors only rise. Ceilings only fall via `gauntlet ratchet`. Fulfillment’s `30` is the Go precedent. |
| Experiment (**H**) | Product CI is the CRAP **constraint** only. Uncle Bob’s experiment stays linked docs. No HTW recreation. No mutation-in-PR in wave 1. |

## Context

Notion task part 2. Adopt Uncle Bob’s CRAP constraint and lessons from [negative-test-experiment](https://github.com/unclebob/negative-test-experiment). Do not invent CRAP variants.

Formula (already in `internal/crap/crap.go`):

`CRAP = CC² × (1 − cov)³ + CC`

Today only `troute-fulfillment` gates CRAP (`crap4go`, ceiling 30, `--changed origin/develop`).

## Language path for the locked set

| Repo | Lang | Next step |
|---|---|---|
| `troute-fulfillment` | Go | Cloud parity (tools on VM = CI) |
| `clark-agency` | Python | `crap4py` after packaging |
| `troute-comms` | Rust | Coverage exists. Adapter after Python |
| `purely-expo` | TypeScript | Coverage first, then `crap4ts` |
| `troute-mcp` | PHP | Coverage first, then `crap4php` |
| Future repos | varies | `templates/CONSUME.md` |

## Scope

**In.** Same formula. Required gate on the locked repos. Cloud-native install. `gauntlet` as the lever. Linked experiment docs.

**Out.** HTW recreation. Mutation as merge gate. SCRAP / hybrid metrics. Ceiling `4` overnight on brownfield. Part 1 skills-pack work.

## Chosen approach

Lever in `gauntlet` + reusable Action + language adapters + Cursor install snippets. Not Makefile copy-paste forever. Not “coverage + complexity lint ≈ CRAP.”

## Phases

1. [phase-1-baseline-and-contract.md](phase-1-baseline-and-contract.md)
2. [phase-2-gauntlet-cloud-packaging.md](phase-2-gauntlet-cloud-packaging.md)
3. [phase-3-fulfillment-cloud-parity.md](phase-3-fulfillment-cloud-parity.md)
4. [phase-4-go-consumer-pilot.md](phase-4-go-consumer-pilot.md)
5. [phase-5-python-crap4py.md](phase-5-python-crap4py.md)
6. [phase-6-rust-troute-comms.md](phase-6-rust-troute-comms.md)
7. [phase-7-ts-purely-expo.md](phase-7-ts-purely-expo.md)
8. [phase-8-php-troute-mcp.md](phase-8-php-troute-mcp.md)
9. [phase-8-optional-experiment-notebook.md](phase-8-optional-experiment-notebook.md) (docs only; Q3 H)
10. [testing.md](testing.md)

## Verification

Planted high-CRAP function fails. Fix passes. Cloud VM runs the same command after `environment.json` install. CI required check. Formula golden tests unchanged.

## Implementation guidance

- `how` before each language adapter.
- `/deslop` before commit. `unslop` on docs/PRs.
- One repo green before the next (**sequence-verifiable-units**).
- Explicit land ask before merge.
