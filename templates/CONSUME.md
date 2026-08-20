# Add CRAP to a repo

How to install the gauntlet CRAP gate in a new repo. Go is wired end-to-end today. Python, Rust, TypeScript, and PHP have scorer recipes in section 6 once their adapters merge. Every scorer uses the same formula in [docs/CRAP.md](../docs/CRAP.md).

## 1. Install the tools

Pin both binaries at `v0.1.0`.

```bash
go install github.com/jadenmaciel/gauntlet/cmd/crap4go@v0.1.0
go install github.com/jadenmaciel/gauntlet/cmd/gauntlet@v0.1.0
```

On Cursor Cloud, call [templates/cursor/cloud-install-go.sh](cursor/cloud-install-go.sh) from `.cursor/environment.json` `install`. The default install directory is `$HOME/.local/bin`. Set `GAUNTLET_BIN_DIR=bin` to match a repo `bin/` directory.

## 2. Set the ceiling from a measured baseline

Create `.gauntlet/thresholds.yml` if the file is missing.

```bash
gauntlet init
```

`gauntlet init` writes `crap_ceiling.value: 8`. Policy E for a brownfield tree is different. Measure first. Write that measured maximum as the first ceiling. After that, only `gauntlet ratchet` may move it, and only tighter.

```yaml
metrics:
  crap_ceiling:
    direction: max
    value: 30
```

`troute-fulfillment` ships at 30 on changed files since `origin/develop`. Use 30 when your measured max is 30. Otherwise write the number you measured.

```bash
go test -coverprofile=coverage.out ./...
crap4go --dir . --profile coverage.out --format json
```

A non-zero exit is expected until the ceiling exists. Read `summary.max_crap` and put that number in `crap_ceiling.value`. Later, tighten with:

```bash
gauntlet ratchet --metric crap_ceiling --value <new-lower-number>
```

## 3. Gate locally

Copy the pieces you need from [templates/go/Makefile.fragment](go/Makefile.fragment). Wire `crap` into the existing `check` target.

Fetch the base branch before `--changed`. A shallow clone without `origin/<base>` fails `git merge-base`.

## 4. Gate in GitHub Actions

After tests write `coverage.out`, call the composite Action.

```yaml
- uses: jadenmaciel/gauntlet/.github/actions/crap-go@v0.1.0
  with:
    profile: coverage.out
    ceiling-file: .gauntlet/thresholds.yml
    changed-ref: origin/main
```

The job must already have Go on `PATH`. The Action installs pinned `crap4go`, reads `metrics.crap_ceiling.value` with the same awk as the Makefile fragment, fetches `changed-ref` when that ref is missing, and fails when any function fails.

Leave `changed-ref` empty to score the whole tree.

## 5. Cursor Cloud

1. Run the install script in `environment.json` `install`.
2. Fetch the base ref in the same install, or before the first `--changed` run.
3. Run the same `crap4go` command CI runs.

Cloud is done when those three match CI, including a failing function failing the job.

## 6. Non-Go adapter recipes

Only Go has a composite GitHub Action today. Each language below ships a Makefile fragment and scorer CLI in its adapter PR. Copy the fragment for repo wiring, generate real test coverage, then run the scorer. Policy E from section 2 applies: measure `summary.max_crap` on real coverage before the ceiling is a gate. Policy H in [docs/uncle-bob-negative-test-experiment.md](../docs/uncle-bob-negative-test-experiment.md) is linked documentation, not a product gate.

### Python (`crap4py`)

Use [templates/python/Makefile.fragment](python/Makefile.fragment).

Install runtime tooling:

```bash
python3 -m pip install coverage
```

Generate coverage and score your repo (after vendoring `adapters/python/crap4py.py`):

```bash
python3 -m coverage run -m pytest
python3 -m coverage xml -o coverage.xml
python3 adapters/python/crap4py.py --dir . --coverage coverage.xml --thresholds .gauntlet/thresholds.yml --format json
```

### Rust (`crap4rs`)

Use [templates/rust/Makefile.fragment](rust/Makefile.fragment).

Set `GAUNTLET_VERSION` to a release that contains `crap4rs`, then run the fragment `crap4rs-tools` target. `cargo install --root` uses the repo root so the binary lands at `bin/crap4rs`:

```bash
export GAUNTLET_VERSION=<release-containing-crap4rs>
make crap4rs-tools
```

Generate coverage and score your repo root (same flags as fragment `crap-rust`, with `--format json` to read `summary.max_crap`):

```bash
cargo llvm-cov --json --output-path target/llvm-cov.json
bin/crap4rs --dir . --coverage-json target/llvm-cov.json --ceiling "$(awk '/^  crap_ceiling:/{found=1; next} found && /^    value:/{print $2; exit}' .gauntlet/thresholds.yml)" --format json
```

### TypeScript (`crap4ts`)

Use [templates/typescript/Makefile.fragment](typescript/Makefile.fragment).

Build the adapter with the fragment `gauntlet-ts-tools` target, generate real coverage, then score from your repo root:

```bash
npm run test:coverage
node .tools/gauntlet/adapters/typescript/crap4ts/dist/cli.js --dir . --coverage coverage/coverage-final.json --thresholds .gauntlet/thresholds.yml --format json
```

### PHP (`crap4php`)

Use [templates/php/Makefile.fragment](php/Makefile.fragment).

The fragment assumes you vendored the adapter under `tools/crap4php`. Install it, generate PHPUnit clover coverage, then score:

```bash
composer install --working-dir tools/crap4php --no-interaction --prefer-dist
vendor/bin/phpunit --coverage-clover build/coverage/clover.xml
php tools/crap4php/bin/crap4php --dir . --coverage build/coverage/clover.xml --thresholds .gauntlet/thresholds.yml --format json
```
