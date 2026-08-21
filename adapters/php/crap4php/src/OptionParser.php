<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class OptionParser
{
    /**
     * @param list<string> $args
     */
    public function parse(array $args): Options
    {
        $dir = '.';
        $coverage = null;
        $ceiling = null;
        $thresholds = '.gauntlet/thresholds.yml';
        $format = 'text';
        $changed = '';

        for ($i = 0; $i < count($args); $i++) {
            $arg = $args[$i];
            if ($arg === '--help' || $arg === '-h') {
                throw new HelpRequested();
            }

            if ($arg === '--dir') {
                $dir = $this->nextValue($args, ++$i, '--dir');
                continue;
            }
            if ($arg === '--coverage') {
                $coverage = $this->nextValue($args, ++$i, '--coverage');
                continue;
            }
            if ($arg === '--ceiling') {
                $value = $this->nextValue($args, ++$i, '--ceiling');
                if (!is_numeric($value)) {
                    throw new InvalidOption(sprintf('--ceiling must be numeric, got %s', $value));
                }
                $ceiling = (float) $value;
                continue;
            }
            if ($arg === '--thresholds') {
                $thresholds = $this->nextValue($args, ++$i, '--thresholds');
                continue;
            }
            if ($arg === '--changed') {
                $changed = $this->nextValue($args, ++$i, '--changed');
                continue;
            }
            if ($arg === '--format') {
                $format = $this->nextValue($args, ++$i, '--format');
                continue;
            }

            throw new InvalidOption(sprintf('unknown argument: %s', $arg));
        }

        if ($format !== 'text' && $format !== 'json') {
            throw new InvalidOption(sprintf('invalid --format %s, expected text or json', $format));
        }

        return new Options(
            dir: $dir,
            coverage: $coverage,
            ceiling: $ceiling,
            thresholds: $thresholds,
            format: $format,
            changed: $changed,
        );
    }

    /**
     * @param list<string> $args
     */
    private function nextValue(array $args, int $index, string $option): string
    {
        if (!array_key_exists($index, $args)) {
            throw new InvalidOption(sprintf('missing value for %s', $option));
        }

        return $args[$index];
    }
}

final class InvalidOption extends \RuntimeException
{
}

final class HelpRequested extends \RuntimeException
{
}
