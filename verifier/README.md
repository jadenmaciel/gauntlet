# crap4py verifier artifacts

Independent verification of the Python `crap4py` CRAP adapter.

## parity/main.go
Computes all 13 golden vectors by calling the LOCKED Go reference
`internal/crap.ComputeCRAP` directly, so the Python adapter's golden table
can be cross-checked against the actual Go formula (not just its own
hardcoded expectations).

Run:
    go run ./verifier/parity

## Evidence captured (2026-08-20)
- `python3 -m unittest discover -s adapters/python/tests -v` -> 3 tests OK
- All 13 golden vectors: python == go == expected == independent within 1e-9
- Planted demo FAIL (coverage-low.xml): CRAP=110 > ceiling 30, exit 1
- Planted demo PASS (coverage-high.xml): CRAP=10 <= ceiling 30, exit 0
- Join: function absent from coverage kept at coverage 0, still scored (not dropped)
- `--ceiling` override and thresholds.yml read both exercised via CLI
- Diff touches only adapters/python/** and templates/python/**
