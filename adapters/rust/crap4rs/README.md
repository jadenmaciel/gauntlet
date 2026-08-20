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
  --coverage-json <path-to-repo>/target/llvm-cov.json \
  --format json
```

By default, ceiling comes from `<path-to-repo>/.gauntlet/thresholds.yml` at `metrics.crap_ceiling.value`.
Use `--ceiling <n>` to override.
