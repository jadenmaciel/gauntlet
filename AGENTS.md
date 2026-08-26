# AGENTS.md

## Using this tool on another repo

Read **[`SKILL.md`](SKILL.md)** and **[`adapters.yml`](adapters.yml)**. Those two files are the adoption contract.

## Working on this repo

Read **[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)** for the five invariants and verification loop. Read **[`CONTEXT.md`](CONTEXT.md)** for nouns. Read **[`adapters.yml`](adapters.yml)** for commands — not this file.

### Run the checks

```bash
go test ./...
python3 -m unittest adapters.python.tests.test_crap4py
cargo test --manifest-path adapters/rust/crap4rs/Cargo.toml
npm --prefix adapters/typescript/crap4ts test
composer --working-dir adapters/php/crap4php test

python3 scripts/check-docs-drift.py
python3 scripts/check-import-graph.py
python3 scripts/check-no-placeholder-ceiling.py
python3 scripts/check-cli-contract.py
```

Self-gate on changed Go functions (after coverage):

```bash
go test -coverprofile=coverage.out ./...
go run ./cmd/crap4go --dir . --coverage coverage.out --thresholds .gauntlet/thresholds.yml --changed origin/main
```

### Adding a language

Follow the checklist in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) rule 1: `adapters.yml` first, then scorer, composite action, template, docs, CI job.
