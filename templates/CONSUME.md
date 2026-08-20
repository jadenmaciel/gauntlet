# Add CRAP to a repo

How to install the gauntlet CRAP gate in a new repo. Go works today. Python, Rust, TypeScript, and PHP adapters land in later phases. The score they will compute is the same formula in [docs/CRAP.md](../docs/CRAP.md).

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

## Later languages

Keep the published formula while you wait. Python, Rust, TypeScript, and PHP adapters land in later phases.

| Repo | Language | Until the adapter exists |
|---|---|---|
| `clark-agency` | Python | Keep coverage. Add `crap4py` in a later phase. |
| `troute-comms` | Rust | Coverage exists. Adapter after Python. |
| `purely-expo` | TypeScript | Real coverage first, then `crap4ts`. |
| `troute-mcp` | PHP | Real coverage first, then `crap4php`. |

A future repo in another language follows this file once its adapter ships. Until then, install `gauntlet` for thresholds only.
