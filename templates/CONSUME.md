# Add the CRAP gate to a repo

Step-by-step adoption for one repo. All five languages are wired end-to-end: each has a
scorer with the same flags, a composite GitHub Action, and a Makefile fragment.

For the commands alone, read [`adapters.yml`](../adapters.yml). For the reasoning and
troubleshooting, read [`SKILL.md`](../SKILL.md). The metric contract is
[`docs/CRAP.md`](../docs/CRAP.md).

## 1. Install the ratchet CLI

`gauntlet` manages the threshold file. It is a Go binary regardless of the language you gate.

```bash
go install github.com/jadenmaciel/gauntlet/cmd/gauntlet@v0.2.0
```

On Cursor Cloud, call [`templates/cursor/cloud-install-go.sh`](cursor/cloud-install-go.sh) from
`.cursor/environment.json` `install`. It installs into `$HOME/.local/bin`, or into
`GAUNTLET_BIN_DIR` if you set it.

Then install the scorer for your language — the `install` line from
[`adapters.yml`](../adapters.yml), or the per-language block in
[`README.md`](../README.md#per-language-setup).

## 2. Measure before you gate

This is the step people skip, and skipping it is why a new gate fails on its first run
against code nobody touched.

Generate real coverage, then score with a ceiling nothing can exceed:

```bash
go test -coverprofile=coverage.out ./...
crap4go --dir . --coverage coverage.out --ceiling 100000 --format json
```

Read `summary.max_crap` from that output. That number is your starting ceiling:

```bash
gauntlet init --crap-ceiling <summary.max_crap>
```

The result:

```yaml
metrics:
  crap_ceiling:
    direction: max
    value: 47
```

`gauntlet init` **without** `--crap-ceiling` writes `8` and marks it `PLACEHOLDER` in the file.
That is a refusal to guess, not a recommendation — very few existing codebases start under 8.

From here the ceiling may only fall, and only through the CLI:

```bash
gauntlet ratchet --metric crap_ceiling --value 42
gauntlet verify  --metric crap_ceiling --value 42
```

## 3. Gate locally

Copy your language's fragment into the repo Makefile and wire `crap` into the existing `check`
target:

| language | fragment |
|---|---|
| Go | [`templates/go/Makefile.fragment`](go/Makefile.fragment) |
| Python | [`templates/python/Makefile.fragment`](python/Makefile.fragment) |
| Rust | [`templates/rust/Makefile.fragment`](rust/Makefile.fragment) |
| TypeScript | [`templates/typescript/Makefile.fragment`](typescript/Makefile.fragment) |
| PHP | [`templates/php/Makefile.fragment`](php/Makefile.fragment) |

Each fragment defines two targets: one that scores `--changed origin/$(BASE_BRANCH)` for the
day-to-day gate, and an `-all` variant that scores the whole tree.

Fetch the base branch before using `--changed`. A shallow clone without `origin/<base>` fails
`git merge-base`:

```bash
git fetch --no-tags origin main
```

## 4. Gate in GitHub Actions

After the test step writes a coverage report, call the composite action for the language.

```yaml
- uses: jadenmaciel/gauntlet/.github/actions/crap-go@v0.2.0
  with:
    coverage: coverage.out
    thresholds: .gauntlet/thresholds.yml
    changed-ref: origin/main
```

| language | action | toolchain the job needs first |
|---|---|---|
| Go | `crap-go` | `actions/setup-go` |
| Python | `crap-python` | `actions/setup-python` |
| Rust | `crap-rust` | a Rust toolchain, e.g. `dtolnay/rust-toolchain` |
| TypeScript | `crap-typescript` | `actions/setup-node` |
| PHP | `crap-php` | `shivammathur/setup-php` |

Every action takes the same inputs: `coverage`, `dir`, `thresholds`, `ceiling`, `changed-ref`,
`format`, `working-directory`, `version`. Each installs its scorer, fetches `changed-ref` when
the ref is missing from a shallow clone, and fails the job when any function is over the
ceiling. Leave `changed-ref` empty to score the whole tree.

`ceiling` overrides `thresholds` when both are set. Earlier versions of the Go action scraped
`metrics.crap_ceiling.value` out of the YAML with `awk`; every scorer now reads the file itself,
so nothing needs to.

## 5. Cursor Cloud

1. Run the install script from `environment.json` `install`.
2. Fetch the base ref in the same install step, or before the first `--changed` run.
3. Run the identical scorer command CI runs.

Cloud is done when all three match CI, including a failing function failing the job.

## 6. What to expect the first week

- **The gate passes but scores nothing.** `--changed` found an empty diff. It warns on stderr;
  check the base ref.
- **Everything reports 0% coverage.** The coverage command did not run, wrote elsewhere, or —
  for PHP — ran without Xdebug/PCOV. Run it on its own and inspect the report first.
- **Exit 2 about the ceiling.** Neither `--ceiling` nor a readable `--thresholds` resolved.
  No scorer invents a default.

Fuller list in [`SKILL.md`](../SKILL.md#troubleshooting).

## 7. Where the ceiling should end up

Do not aim for 4. The research this formula comes from found that driving CRAP down that far
raises coverage and hurts readability without improving design — see
[`docs/uncle-bob-negative-test-experiment.md`](../docs/uncle-bob-negative-test-experiment.md).
Start at your measured baseline, ratchet down when a real refactor earns it, and let review
carry the design argument.
