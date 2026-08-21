# AGENTS.md

## Using this tool on another repo

**[`SKILL.md`](SKILL.md) is canonical.** Read it, and read
[`adapters.yml`](adapters.yml) for the exact per-language commands. Everything an agent needs to
add the CRAP gate to a Go, Python, Rust, TypeScript, or PHP repo is in those two files. This
file exists so agents that look for `AGENTS.md` by convention find their way there, and to hold
the notes that only apply to working *on* gauntlet itself.

Short version: read `adapters.yml`, detect the language, measure a baseline before setting a
ceiling, generate coverage, score, and only ever ratchet the ceiling down.

## Working on this repo

Five scorers implement one metric. They must stay interchangeable.

### Layout

| path | what |
|---|---|
| `internal/crap/` | the formula and the changed-file filter — the reference implementation |
| `internal/thresholds/` | ratchet file load/save/verify |
| `cmd/crap4go/`, `cmd/gauntlet/` | Go scorer and the ratchet CLI |
| `adapters/{python,rust,typescript,php}/` | the other four scorers |
| `adapters.yml` | the manifest; docs must agree with it |
| `.github/actions/crap-*/` | one composite action per language |
| `templates/` | copy-paste Makefile fragments and the adoption guide |

### Run the tests

```bash
go test ./...
python3 -m unittest adapters.python.tests.test_crap4py
cargo test --manifest-path adapters/rust/crap4rs/Cargo.toml
npm --prefix adapters/typescript/crap4ts test
composer --working-dir adapters/php/crap4php test
```

### Rules that are easy to break

- **The CLI surface is a contract.** All five accept
  `--dir --coverage --thresholds --ceiling --changed --format`. Adding a flag to one means
  adding it to all five, updating `adapters.yml`, and updating `README.md` + `SKILL.md` — the
  `docs-drift` CI job fails otherwise.
- **The JSON shape is a contract.** `ceiling`, `functions[]`, `summary` with the keys listed in
  `adapters.yml`. Consumers parse it.
- **Exit codes are a contract.** `0` pass, `1` over ceiling, `2` usage or I/O error.
- **Never invent a ceiling.** No scorer may default one. `--ceiling` → `--thresholds` → exit 2.
- **Never let a missing input read as a passing result.** A missing coverage report exits 2. An
  empty `--changed` diff passes but warns on stderr.
- **Git fixtures must be hermetic.** Every test that runs `git` sets
  `GIT_CONFIG_GLOBAL=/dev/null` and `GIT_CONFIG_SYSTEM=/dev/null`. A developer's global
  `core.hooksPath` can write generated files into a fixture on commit, which then show up as
  untracked changes and make `--changed` assertions flaky.
- **Temp fixture names need more than a clock.** Tests run in parallel and `SystemTime` is not
  nanosecond-resolution on macOS, so two fixtures collide. Use the process id plus a counter.

### Adding a language

1. Add the entry to `adapters.yml` first — it is what the drift check reads.
2. Implement the six flags and the shared JSON shape.
3. Add a composite action under `.github/actions/crap-<lang>/`.
4. Add a `templates/<lang>/Makefile.fragment`.
5. Add per-language blocks to `README.md` and `SKILL.md` using the exact strings from the
   manifest.
6. Add a CI job.
