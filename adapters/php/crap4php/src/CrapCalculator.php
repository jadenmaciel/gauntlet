<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class CrapCalculator
{
    public static function compute(int $complexity, float $coveragePercent): float
    {
        $cc = (float) $complexity;
        $coverage = $coveragePercent / 100.0;
        $uncovered = 1.0 - $coverage;

        return $cc * $cc * $uncovered * $uncovered * $uncovered + $cc;
    }
}
