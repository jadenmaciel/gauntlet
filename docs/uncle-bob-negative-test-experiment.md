# Uncle Bob negative-test experiment

About the research this repo cites, and what product CI does with it.

## Source

[unclebob/negative-test-experiment](https://github.com/unclebob/negative-test-experiment)

Eight independent Hunt the Wumpus programs. Testing discipline crossed with CRAP below 4, then mutation.

## Conclusion we cite

`experiment-summary.md` in that repo:

> CRAP raises coverage and wrecks cleanliness; it does not improve design.

## What we copy

- The published formula, already in `internal/crap/crap.go`
- CRAP as a constraint that fails the build when a function is above the ceiling
- The cleanliness warning. Product ceilings start from a measured baseline (policy E), not 4

## What we leave

- Recreating the eight-cell Hunt the Wumpus grid
- Forcing CRAP below 4 on product trees
- Mutation as a merge requirement on pull requests

Those stay in the upstream notebooks. Product CI does not run them.

## Policy H

The experiment is linked documentation. It is not a product gate.

Product adoption of the CRAP constraint is [docs/CRAP.md](CRAP.md).
