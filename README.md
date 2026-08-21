# gauntlet

Find the functions most likely to break, and stop new ones from being added.

gauntlet scores every function with CRAP (Change Risk Anti-Pattern). That score
combines how branchy a function is with how well it is tested. The build fails when
any function is over a ceiling. The ceiling lives in a file and may only fall, so
the bar never slips.

One metric, five languages: Go, Python, Rust, TypeScript, PHP. Same flags, same JSON,
same exit codes.

> Pointing an AI agent at this repo? Send it to [SKILL.md](SKILL.md) and
> [adapters.yml](adapters.yml). See [For AI agents](#for-ai-agents).

## Quick start

```bash
# 1. Install the ratchet CLI.
go install github.com/jadenmaciel/gauntlet/cmd/gauntlet@v0.2.0

# 2. Install your language's scorer, generate coverage, and measure the baseline.
#    (Substitute your language's three lines from the table below.)
go install github.com/jadenmaciel/gauntlet/cmd/crap4go@v0.2.0
go test -coverprofile=coverage.out ./...
crap4go --dir . --coverage coverage.out --ceiling 100000 --format json | jq .summary

# 3. Start the ratchet at the number you just measured, not at a guess.
gauntlet init --crap-ceiling 47

# 4. From now on this fails the build when a function is over the ceiling.
crap4go --dir . --coverage coverage.out --thresholds .gauntlet/thresholds.yml
```

Step 2 matters. A ceiling picked without measuring either fails on day one against
code nobody just wrote, or sits so high it gates nothing.

## Languages

| Language | Scorer | Runtime floor | Coverage format | Coverage command |
|---|---|---|---|---|
| Go | `crap4go` | Go 1.26.4 | `go test` cover profile | `go test -coverprofile=coverage.out ./...` |
| Python | `crap4py` | Python 3.10 | Cobertura XML | `python3 -m coverage run -m pytest && python3 -m coverage xml -o coverage.xml` |
| Rust | `crap4rs` | Rust stable | `cargo-llvm-cov` JSON | `cargo llvm-cov --json --output-path target/llvm-cov.json` |
| TypeScript | `crap4ts` | Node 20.6 | istanbul JSON or lcov | `npx c8 --reporter=json npm test` |
| PHP | `crap4php` | PHP 8.2 | PHPUnit Clover XML | `vendor/bin/phpunit --coverage-clover coverage.xml` |

Every scorer takes the same flags:

```
<scorer> [--dir <path>] [--coverage <file>] [--thresholds <file>]
         [--ceiling <number>] [--changed <ref>] [--format text|json]
```

| flag | meaning | default |
|---|---|---|
| `--dir` | root to scan for source files | `.` |
| `--coverage` | coverage report to read | per language, see the table above |
| `--thresholds` | YAML holding `metrics.crap_ceiling.value` | `.gauntlet/thresholds.yml` |
| `--ceiling` | numeric override; wins over `--thresholds` | unset |
| `--changed` | git ref; score only files changed since it | unset = whole tree |
| `--format` | `text` or `json` | `text` |

Ceiling precedence is `--ceiling`, then the `--thresholds` file, then exit 2.
No scorer invents a ceiling.

Exit codes are `0` pass, `1` at least one function over the ceiling, and `2` for a
usage or I/O error.

## Per-language setup

Three lines each: install, cover, score.

### Go

```bash
go install github.com/jadenmaciel/gauntlet/cmd/crap4go@v0.2.0
go test -coverprofile=coverage.out ./...
crap4go --dir . --coverage coverage.out --thresholds .gauntlet/thresholds.yml
```

`--profile` still works as an alias for `--coverage`.

### Python

```bash
pip install "git+https://github.com/jadenmaciel/gauntlet@v0.2.0#subdirectory=adapters/python"
python3 -m coverage run -m pytest && python3 -m coverage xml -o coverage.xml
crap4py --dir . --coverage coverage.xml --thresholds .gauntlet/thresholds.yml
```

`crap4py.py` uses only the standard library, so vendoring the single file works too.
See [adapters/python/README.md](adapters/python/README.md).

### Rust

```bash
cargo install --git https://github.com/jadenmaciel/gauntlet --tag v0.2.0 crap4rs
cargo llvm-cov --json --output-path target/llvm-cov.json
crap4rs --dir . --coverage target/llvm-cov.json --thresholds .gauntlet/thresholds.yml
```

Install the coverage tool first with `rustup component add llvm-tools-preview` and
`cargo install cargo-llvm-cov --locked`. `--coverage-json` and `--ceiling-file` remain as
aliases. See [adapters/rust/crap4rs/README.md](adapters/rust/crap4rs/README.md).

### TypeScript

```bash
git clone --depth 1 --branch v0.2.0 https://github.com/jadenmaciel/gauntlet .tools/gauntlet && npm --prefix .tools/gauntlet/adapters/typescript/crap4ts ci && npm --prefix .tools/gauntlet/adapters/typescript/crap4ts run build
npx c8 --reporter=json npm test
node .tools/gauntlet/adapters/typescript/crap4ts/dist/cli.js --dir . --coverage coverage/coverage-final.json --thresholds .gauntlet/thresholds.yml
```

Any tool that emits istanbul `coverage-final.json` or `lcov.info` works (jest, vitest, nyc, c8).
See [adapters/typescript/crap4ts/README.md](adapters/typescript/crap4ts/README.md).

### PHP

```bash
git clone --depth 1 --branch v0.2.0 https://github.com/jadenmaciel/gauntlet tools/gauntlet && composer install --working-dir tools/gauntlet/adapters/php/crap4php --no-interaction
vendor/bin/phpunit --coverage-clover coverage.xml
php tools/gauntlet/adapters/php/crap4php/bin/crap4php --dir src --coverage coverage.xml --thresholds .gauntlet/thresholds.yml
```

Clover XML needs Xdebug or PCOV. Without a driver, PHPUnit writes an empty report and every
function scores as untested. See [adapters/php/crap4php/README.md](adapters/php/crap4php/README.md).

## Output

`--format text` is for reading:

```
src/hot.ts:12:hotPath	complexity=10	coverage=10.0%	crap=82.90	FAIL

ceiling=30.00 total=1 failing=1 max_crap=82.90
```

`--format json` is for tooling. The keys and structure are identical across all five scorers:

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

The two numbers worth reading in a script are `summary.max_crap` and `summary.failing`.

Numeric fields are JSON numbers, so a whole value may render as `30` or `30.0` depending on the
scorer. Parse the JSON. Do not diff the raw text.

## Thresholds and ratcheting

`.gauntlet/thresholds.yml` holds floors that may only rise and ceilings that may only fall:

```yaml
metrics:
  crap_ceiling:
    direction: max
    value: 30
```

Create it from a measured baseline:

```bash
gauntlet init --crap-ceiling 47
```

Then tighten it, never by hand:

```bash
gauntlet ratchet --metric crap_ceiling --value 42
gauntlet verify  --metric crap_ceiling --value 42
```

`gauntlet init` without `--crap-ceiling` writes `8` and marks it `PLACEHOLDER` in the file.
That is not a recommended starting ceiling. Most existing codebases start well above 8, and
gating there fails on day one. Measure first, then pass `--crap-ceiling`.

## CI

Score only what the pull request touched so the gate applies to new code without a
tree-wide cleanup first:

```bash
crap4go --dir . --coverage coverage.out --thresholds .gauntlet/thresholds.yml --changed origin/main
```

`--changed` resolves the merge base with the ref and adds untracked files. If the diff is
empty, the scorer passes but warns on stderr, so a misconfigured base ref cannot silently
disable the gate.

There is a composite action per language:

```yaml
- uses: jadenmaciel/gauntlet/.github/actions/crap-go@v0.2.0
  with:
    coverage: coverage.out
    thresholds: .gauntlet/thresholds.yml
    changed-ref: origin/main
```

Also `crap-python`, `crap-rust`, `crap-typescript`, `crap-php`. Each installs its scorer, fetches
the base ref when the clone is shallow, and fails the job on exit 1. Copy-paste Makefile targets
live in [templates/](templates/). Full adoption steps are in
[templates/CONSUME.md](templates/CONSUME.md).

## For AI agents

[SKILL.md](SKILL.md) is the decision procedure, in order, with troubleshooting. Start there.

[adapters.yml](adapters.yml) is the machine-readable manifest. It has per-language detect globs,
install, coverage command, coverage format, and score command. Read it instead of parsing
prose. CI fails if the docs and the manifest disagree.

[AGENTS.md](AGENTS.md) points at both, plus the contracts to preserve when changing this repo.

Give an agent the repo URL and "add the CRAP gate to this project." Those three files are enough.

## The CRAP formula

```
CRAP = CC^2 * (1 - cov)^3 + CC
```

`CC` is cyclomatic complexity. `cov` is that function's statement coverage from 0.0 to 1.0.

Coverage is cubed, so tests move the score much faster than refactoring does:

| complexity | coverage | CRAP |
|---|---|---|
| 10 | 0% | 110 |
| 10 | 50% | 22.5 |
| 10 | 90% | 10.1 |
| 10 | 100% | 10 |

| CRAP | Reading |
|---|---|
| 1-5 | clean |
| 5-30 | moderate risk |
| 30+ | crappy. Refactor or test before extending it. |

## What counts as complexity

Each function starts at 1, then adds one for every branch point. Nested functions are counted
separately rather than folded into their parent. The list below is the union across the five
languages. Each scorer counts the constructs its language actually has.

- `if` / `else if`, ternaries and conditional expressions
- every loop (`for`, `while`, `foreach`, `range`, comprehensions, plus one per comprehension `if`)
- each non-default `case` / `match` arm, and each `select` communication clause
- each `catch` / `except` handler
- each `&&` / `||` / `and` / `or` operand past the first
- `assert` (Python), where a failure is a second path

## The honest caveat

The formula comes from [unclebob/negative-test-experiment](https://github.com/unclebob/negative-test-experiment),
whose own conclusion is worth quoting:

> CRAP raises coverage and wrecks cleanliness; it does not improve design.

Pushing a ceiling to 4 buys tests written to satisfy arithmetic. Use CRAP to stop new
high-score functions from landing. Measure a baseline, tighten deliberately, and leave design
judgement to review. See
[docs/uncle-bob-negative-test-experiment.md](docs/uncle-bob-negative-test-experiment.md).

## Development

```bash
go test ./...
python3 -m unittest adapters.python.tests.test_crap4py
cargo test --manifest-path adapters/rust/crap4rs/Cargo.toml
npm --prefix adapters/typescript/crap4ts test
composer --working-dir adapters/php/crap4php test
```

The contracts that must not drift (flags, JSON shape, exit codes) are listed in
[AGENTS.md](AGENTS.md). The product contract is [docs/CRAP.md](docs/CRAP.md).

Known gap: upstream `crap4go` runs the tests itself. gauntlet requires you to generate
coverage first. The three-line blocks above are the workaround, not the fix.

## License

[MIT](LICENSE).

Descended from [unclebob/crap4go](https://github.com/unclebob/crap4go).
