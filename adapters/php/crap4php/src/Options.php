<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class Options
{
    public function __construct(
        public readonly string $dir,
        public readonly ?string $coverage,
        public readonly ?float $ceiling,
        public readonly string $thresholds,
        public readonly string $format,
        public readonly string $changed,
    ) {
    }
}
