<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class Summary
{
    public function __construct(
        public readonly int $total,
        public readonly int $failing,
        public readonly float $maxCrap,
    ) {
    }
}
