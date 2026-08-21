# crap4php

PHP CRAP scorer for gauntlet.

## Formula

`CRAP = CC^2 * (1 - cov)^3 + CC`, where `cov = coverage_percent / 100`.

## Requirements

- PHP >= 8.2 with the `simplexml` extension.
- Composer, for `nikic/php-parser` and `symfony/yaml`.

## Install

```bash
composer require --dev jadenmaciel/crap4php
```

Until the package is published, install from a checkout:

```bash
git clone --depth 1 --branch v0.2.0 https://github.com/jadenmaciel/gauntlet tools/gauntlet
composer install --working-dir tools/gauntlet/adapters/php/crap4php --no-interaction
```

## Score a PHP project

```bash
vendor/bin/phpunit --coverage-clover build/coverage/clover.xml
php tools/gauntlet/adapters/php/crap4php/bin/crap4php \
  --dir src --coverage build/coverage/clover.xml --thresholds .gauntlet/thresholds.yml
```

Coverage format: PHPUnit Clover XML. Generating it needs Xdebug or PCOV enabled.

## Flags

| flag | meaning | default |
|---|---|---|
| `--dir` | root to scan for `.php` files | `.` |
| `--coverage` | Clover XML to read (required) | none |
| `--thresholds` | YAML holding `metrics.crap_ceiling.value` | `.gauntlet/thresholds.yml` |
| `--ceiling` | numeric override; wins over `--thresholds` | unset |
| `--changed` | git ref; score only files changed since it | unset = whole tree |
| `--format` | `text` or `json` | `text` |

Exit codes: `0` pass, `1` at least one function over the ceiling, `2` usage or I/O error.

## Test

```bash
composer install --no-interaction
composer test
```
