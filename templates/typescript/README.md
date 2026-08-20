# TypeScript consume fragment

Copy `templates/typescript/Makefile.fragment` into your repo Makefile and wire `crap` into your existing check target.

`crap4ts` expects a coverage report file. The fragment assumes `coverage/coverage-final.json` from Istanbul-compatible tooling (for example, Jest, Vitest, or nyc).

If your coverage output is lcov, point `COVERAGE_JSON` at your `.info` file.
