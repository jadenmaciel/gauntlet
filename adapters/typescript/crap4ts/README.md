# crap4ts

`crap4ts` computes CRAP scores for TypeScript functions using the locked formula:

`CRAP = CC^2 * (1 - cov)^3 + CC`

`cov` is a fraction from 0 to 1. Internally, the scorer follows the Go reference convention and feeds `coveragePercent` into the formula implementation.

## Coverage input

- Primary: Istanbul `coverage-final.json` (function-level coverage from statement coverage inside function ranges).
- Also accepted: lcov (`.lcov` / `.info`) using `FN`/`FNDA` rows (hit/miss-style function coverage).

Text and JSON reports express coverage as a percentage from 0 to 100.

## Commands

```bash
npm ci
npm test
npm run build
node dist/cli.js --dir . --coverage coverage/coverage-final.json --format json
```
