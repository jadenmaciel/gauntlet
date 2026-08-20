# Testing (cross-cutting)

Back-link: [overview.md](overview.md)

## Static

- Go: `go test ./...` in gauntlet, including golden `ComputeCRAP` cases.
- Adapters: language-native unit tests + shared golden JSON vectors.
- Consumer repos: existing `make check` / CI entrypoints stay green after wiring.

## Runtime

- **CLI.** control-cli against `crap4go` / adapter stdout: planted fail, clean pass, JSON format.
- **CI.** Required check visible on a PR.
- **Cloud.** Fresh Cursor Cloud Agent on a wired repo runs the check after env install (prove base-ref fetch).
- **Browser.** Not primary. Dashboard notifications are not the gate.

## Fixtures

- Keep tiny plant-and-fail fixtures inside gauntlet `testdata/` so reviewers can rerun without product repos.

## Non-goals

- Mutation kill scores as merge blockers.
- Replaying Hunt the Wumpus.
