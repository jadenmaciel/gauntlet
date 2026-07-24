# gremlins unleash — flag reference

Flags relevant to this gauntlet's mutation-testing setup. Not a full
gremlins reference — only the flags this repo's tooling actually uses.

| Flag | Short | Description |
|---|---|---|
| `--diff` | `-D` | Restrict mutation testing to lines changed against a diff/ref, instead of the whole codebase. |
| `--threshold-efficacy` | | Minimum efficacy percentage required to pass; `gremlins unleash` exits non-zero below it. |
| `--threshold-mcover` | | Minimum mutant coverage percentage required to pass; exits non-zero below it. |
| `--output` | `-o` | Path to write the JSON results report. |
| `--exclude-files` | `-E` | Regular expression for files to exclude from mutation. |
| `--timeout-coefficient` | | Multiplier on the baseline test-suite runtime used to compute each mutant's kill timeout. |
| `--integration` | `-i` | Include integration tests (not just unit tests) when running the mutated suite. |

## Efficacy vs. mutant coverage — why both thresholds exist

Gremlins classifies each mutant as one of: **killed**, **lived**,
**not covered**, or **timed out**.

- **Efficacy** = `KILLED / (KILLED + LIVED)`
  Only counts mutants that were actually covered and ran to a verdict.
  Not-covered and timed-out mutants aren't in this ratio at all — a file
  with no test coverage doesn't drag efficacy down.

- **Mutant coverage** = `(KILLED + LIVED) / TOTAL`
  Counts against the full mutant population. Not-covered and timed-out
  mutants sit in the denominator with nothing in the numerator, so they
  do pull this ratio down.

Because not-covered and timed-out mutants affect one ratio and not the
other, a single threshold can't stand in for both: `--threshold-efficacy`
catches "tests that run but don't actually assert anything meaningful,"
while `--threshold-mcover` catches "code with no mutation coverage at
all." Set both.
