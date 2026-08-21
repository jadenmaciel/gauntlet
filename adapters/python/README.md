# crap4py

Python CRAP scorer for gauntlet.

## Formula

`CRAP = CC^2 * (1 - cov)^3 + CC`, where `cov = coverage_percent / 100`.

## Requirements

- Python >= 3.10. This is a hard floor, not a preference: the complexity visitor
  references `ast.Match` / `ast.MatchAs`, which do not exist before 3.10.
- No third-party runtime dependencies. `coverage`/`pytest-cov` are only needed to
  produce the coverage report, not to score it.

## Install

```bash
pip install "git+https://github.com/jadenmaciel/gauntlet@v0.2.0#subdirectory=adapters/python"
```

`crap4py.py` is also deliberately a single file with no imports outside the standard
library, so vendoring it into a repo works just as well:

```bash
curl -sSfLO https://raw.githubusercontent.com/jadenmaciel/gauntlet/v0.2.0/adapters/python/crap4py.py
```

## Score a Python project

```bash
python3 -m coverage run -m pytest
python3 -m coverage xml -o coverage.xml
crap4py --dir . --coverage coverage.xml --thresholds .gauntlet/thresholds.yml
```

Coverage format: Cobertura XML, as written by `coverage xml` or
`pytest --cov --cov-report=xml`.

## Flags

| flag | meaning | default |
|---|---|---|
| `--dir` | root to scan for `.py` files | `.` |
| `--coverage` | Cobertura XML to read (required) | none |
| `--thresholds` | YAML holding `metrics.crap_ceiling.value` | `.gauntlet/thresholds.yml` |
| `--ceiling` | numeric override; wins over `--thresholds` | unset |
| `--changed` | git ref; score only files changed since it | unset = whole tree |
| `--format` | `text` or `json` | `text` |

Exit codes: `0` pass, `1` at least one function over the ceiling, `2` usage or I/O error.

## Test

```bash
python3 -m unittest discover -s adapters/python/tests -t .
```
