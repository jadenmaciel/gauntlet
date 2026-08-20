<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

use Symfony\Component\Yaml\Yaml;

final class Thresholds
{
    public static function readCrapCeiling(string $path): float
    {
        if (!is_file($path)) {
            throw new \RuntimeException(sprintf('thresholds file not found: %s', $path));
        }

        $parsed = Yaml::parseFile($path);
        if (!is_array($parsed)) {
            throw new \RuntimeException(sprintf('invalid thresholds YAML: %s', $path));
        }

        $metrics = $parsed['metrics'] ?? null;
        if (!is_array($metrics)) {
            throw new \RuntimeException('thresholds YAML missing metrics map');
        }

        $crapCeiling = $metrics['crap_ceiling'] ?? null;
        if (!is_array($crapCeiling)) {
            throw new \RuntimeException('thresholds YAML missing metrics.crap_ceiling');
        }

        $direction = $crapCeiling['direction'] ?? null;
        if ($direction !== 'max') {
            throw new \RuntimeException('metrics.crap_ceiling.direction must be "max"');
        }

        $value = $crapCeiling['value'] ?? null;
        if (!is_numeric($value)) {
            throw new \RuntimeException('metrics.crap_ceiling.value must be numeric');
        }

        return (float) $value;
    }
}
