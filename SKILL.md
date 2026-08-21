---
name: gauntlet-crap
description: >
  Computes CRAP (Change Risk Anti-Pattern) scores per function by combining cyclomatic
  complexity with test coverage, and fails a build when any function exceeds a ratcheting
  ceiling. Supports Go, Python, Rust, TypeScript, and PHP. Use when the user asks for a CRAP
  report, a code quality gate, cyclomatic complexity analysis, or wants to add a
  complexity/coverage threshold to CI.
---

# gauntlet-crap

Score every function in a repo, fail the build on the risky ones, and tighten the bar over
time without ever letting it slip back.

**Read [`adapters.yml`](adapters.yml) first.** It is the machine-readable source of truth for
install, coverage, and score commands. This file explains the reasoning; that file holds the
commands.

## The metric

```
CRAP = CC^2 * (1 - cov)^3 + CC
```

- `CC` — cyclomatic complexity: how many independent paths run through the function.
- `cov` — that function's statement coverage, `0.0` to `1.0`.

Coverage is **cubed**, so writing tests moves the score far faster than simplifying code does.
A complexity-10 function at 0% coverage scores 110; at 50% coverage it is already down to 22.5;
at 100% coverage it is 10.

| CRAP | Reading |
|---|---|
| 1–5 | clean |
| 5–30 | moderate risk |
| 30+ | crappy — refactor or test before extending it |

## Decision procedure

Follow these in order. Do not skip step 3.

1. **Read `adapters.yml`.** Take commands from it verbatim rather than from prose.
2. **Detect the language** using each entry's `detect` globs. If several match (a polyglot
   repo), ask the user which tree to gate instead of guessing.
3. **Establish the ceiling before gating.** If `.gauntlet/thresholds.yml` does not exist:
   ```bash
   # Score with a ceiling nothing can exceed, to read the real baseline.
   <scorer> --dir . --coverage <report> --ceiling 100000 --format json
   gauntlet init --crap-ceiling <the summary.max_crap you just read>
   ```
   Never gate an existing codebase at the placeholder. `gauntlet init` without
   `--crap-ceiling` writes `8` and labels it `PLACEHOLDER` in the file, because a real
   codebase almost never starts under 8 and the first run would fail on code nobody touched.
4. **Generate coverage, then score.** Confirm the coverage file exists first — a scorer given
   a missing report exits `2` rather than reporting everything as untested.
5. **Read the JSON.** `summary.max_crap` and `summary.failing` are the two numbers that matter.
6. **On exit 1**, take the highest-CRAP function and either cover its branches or split it.
   Rerun. Prefer tests: coverage is the cubed term.
7. **Tighten only with `gauntlet ratchet`.** `direction: max` means the ceiling may only fall.
   ```bash
   gauntlet ratchet --metric crap_ceiling --value <new-lower-number>
   ```

## The CLI, identical in all five scorers

```
<scorer> [--dir <path>] [--coverage <file>] [--thresholds <file>]
         [--ceiling <number>] [--changed <ref>] [--format text|json]
```

| flag | meaning | default |
|---|---|---|
| `--dir` | root to scan for source files | `.` |
| `--coverage` | coverage report to read | per language; see `adapters.yml` |
| `--thresholds` | YAML holding `metrics.crap_ceiling.value` | `.gauntlet/thresholds.yml` |
| `--ceiling` | numeric override; **wins over `--thresholds`** | unset |
| `--changed` | git ref; score only files changed since it | unset = whole tree |
| `--format` | `text` or `json` | `text` |

**Ceiling precedence:** `--ceiling` → `--thresholds` file → **exit 2**. Nothing is assumed.

**Exit codes:** `0` pass · `1` at least one function over the ceiling · `2` usage or I/O error.

**JSON shape**, stable across all five scorers:

```json
{
  "ceiling": 30,
  "functions": [
    {"file": "src/hot.ts", "line": 12, "func": "hotPath",
     "complexity": 10, "coverage": 10, "crap": 82.9, "pass": false}
  ],
  "summary": {"total": 1, "failing": 1, "max_crap": 82.9}
}
```

## Per-language commands

Three lines each: install, cover, score.

### Go

```bash
go install github.com/jadenmaciel/gauntlet/cmd/crap4go@v0.2.0
go test -coverprofile=coverage.out ./...
crap4go --dir . --coverage coverage.out --thresholds .gauntlet/thresholds.yml
```

### Python

```bash
pip install "git+https://github.com/jadenmaciel/gauntlet@v0.2.0#subdirectory=adapters/python"
python3 -m coverage run -m pytest && python3 -m coverage xml -o coverage.xml
crap4py --dir . --coverage coverage.xml --thresholds .gauntlet/thresholds.yml
```

### Rust

```bash
cargo install --git https://github.com/jadenmaciel/gauntlet --tag v0.2.0 crap4rs
cargo llvm-cov --json --output-path target/llvm-cov.json
crap4rs --dir . --coverage target/llvm-cov.json --thresholds .gauntlet/thresholds.yml
```

Needs `rustup component add llvm-tools-preview` and `cargo install cargo-llvm-cov --locked`
before the coverage command.

### TypeScript

```bash
git clone --depth 1 --branch v0.2.0 https://github.com/jadenmaciel/gauntlet .tools/gauntlet && npm --prefix .tools/gauntlet/adapters/typescript/crap4ts ci && npm --prefix .tools/gauntlet/adapters/typescript/crap4ts run build
npx c8 --reporter=json npm test
node .tools/gauntlet/adapters/typescript/crap4ts/dist/cli.js --dir . --coverage coverage/coverage-final.json --thresholds .gauntlet/thresholds.yml
```

Any tool that emits istanbul `coverage-final.json` or `lcov.info` works — `jest --coverage`,
`vitest --coverage`, `nyc`, `c8`.

### PHP

```bash
git clone --depth 1 --branch v0.2.0 https://github.com/jadenmaciel/gauntlet tools/gauntlet && composer install --working-dir tools/gauntlet/adapters/php/crap4php --no-interaction
vendor/bin/phpunit --coverage-clover coverage.xml
php tools/gauntlet/adapters/php/crap4php/bin/crap4php --dir src --coverage coverage.xml --thresholds .gauntlet/thresholds.yml
```

Clover XML needs Xdebug or PCOV enabled, or PHPUnit writes an empty report.

## Gating a pull request

`--changed <ref>` scores only files touched since the merge base with `<ref>`, which is what
makes the gate usable on a repo whose history predates it: new code is held to the ceiling
while old code is left alone.

```bash
crap4go --dir . --coverage coverage.out --thresholds .gauntlet/thresholds.yml --changed origin/main
```

In GitHub Actions, use the composite action for the language and pass `changed-ref`:

```yaml
- uses: jadenmaciel/gauntlet/.github/actions/crap-go@v0.2.0
  with:
    coverage: coverage.out
    thresholds: .gauntlet/thresholds.yml
    changed-ref: origin/main
```

Equivalents: `crap-python`, `crap-rust`, `crap-typescript`, `crap-php`.

## Troubleshooting

**`git merge-base` fails, or `--changed` errors on CI.**
The clone is shallow and has no common ancestor with the base branch. Fetch it first:
```bash
git fetch --no-tags origin main
```
The composite actions already do this when the ref is missing.

**`no files changed since '<ref>'; nothing was scored`.**
The gate passed because there was nothing to score. That is legitimate, but if you did not
expect it the base ref is wrong — check for `origin/main` vs `main`, or a missing fetch. Before
v0.2.0 this passed silently; the warning exists so a misconfigured gate is visible.

**Every function reports 0% coverage.**
Three causes, in order of likelihood:
1. The coverage command did not actually run, or wrote to a different path. Run it alone and
   confirm the file exists and is non-empty.
2. The coverage report records absolute paths from a different root than `--dir`. This bites on
   macOS, where `/tmp` and `/var` resolve through `/private`, so a path recorded during the test
   run does not literally sit under the scan root. The scorers resolve symlinks on both sides to
   join across it; if you see this on v0.2.0 or later, it is a bug worth reporting.
3. For PHP: no coverage driver. Clover XML needs Xdebug or PCOV; without one PHPUnit emits an
   empty report and everything scores as untested.

**Go: functions in the report do not match files I changed.**
`go tool cover -func` prints module-import paths (`github.com/you/repo/pkg/file.go`), not
repo-relative ones. The scorers match on a path-segment suffix in either direction, so this
normally just works; if you filter the JSON yourself, match on suffix, not equality.

**`exit 2` with a message about the ceiling.**
Neither `--ceiling` nor a readable `--thresholds` file resolved. Either pass `--ceiling <n>` or
create the file with `gauntlet init --crap-ceiling <measured baseline>`. As of v0.2.0 no scorer
invents a default, because a silent default contradicted the committed threshold.

**`--coverage` is required and I do not have a report yet.**
That is deliberate. A missing report used to score every function at 0% coverage, which looks
like a real result. Generate coverage first.

## The honest caveat

From [unclebob/negative-test-experiment](https://github.com/unclebob/negative-test-experiment),
which this repo's formula comes from:

> CRAP raises coverage and wrecks cleanliness; it does not improve design.

Driving a ceiling to 4 produces test suites written to satisfy an arithmetic constraint. Use
CRAP as a **brake on new risk**, not a design target: start from a measured baseline, ratchet
down deliberately, and let human review carry the design argument.

## Related

- [`adapters.yml`](adapters.yml) — the manifest to read programmatically
- [`README.md`](README.md) — human-facing overview
- [`docs/CRAP.md`](docs/CRAP.md) — the internal product contract
- [`templates/CONSUME.md`](templates/CONSUME.md) — step-by-step adoption for a new repo
