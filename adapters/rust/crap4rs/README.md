# crap4rs

Rust CRAP scorer for gauntlet.

## Formula

`CRAP = CC^2 * (1 - cov)^3 + CC`, where `cov = coverage_percent / 100`.

## Install tooling

```bash
rustup component add llvm-tools-preview
cargo install cargo-llvm-cov --version 0.6.21 --locked
```

## Build and test

```bash
cargo test --manifest-path adapters/rust/crap4rs/Cargo.toml
```

## Score a Rust project

Generate coverage JSON first:

```bash
cargo llvm-cov --manifest-path <path-to-repo>/Cargo.toml --json --output-path <path-to-repo>/target/llvm-cov.json
```

Then run the scorer:

```bash
cargo run --manifest-path adapters/rust/crap4rs/Cargo.toml -- \
  --dir <path-to-repo> \
  --coverage <path-to-repo>/target/llvm-cov.json \
  --format json
```

## Flags

| flag | meaning | default |
|---|---|---|
| `--dir` | root to scan | `.` |
| `--coverage` | `cargo llvm-cov --json` output | `target/llvm-cov.json` |
| `--thresholds` | YAML holding `metrics.crap_ceiling.value` | `<dir>/.gauntlet/thresholds.yml` |
| `--ceiling` | numeric override; wins over `--thresholds` | unset |
| `--changed` | git ref; score only files changed since it | unset = whole tree |
| `--format` | `text` or `json` | `text` |

`--coverage-json` and `--ceiling-file` are the pre-v0.2.0 names and still work.

Exit codes: `0` pass, `1` a function is over the ceiling, `2` usage or I/O error. A missing
coverage file or an unresolvable ceiling is exit 2, never a silent pass.
