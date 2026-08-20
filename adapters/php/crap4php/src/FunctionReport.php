<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class FunctionReport
{
    public function __construct(
        public readonly string $file,
        public readonly int $line,
        public readonly string $func,
        public readonly int $complexity,
        public readonly float $coverage,
        public readonly float $crap,
        public readonly bool $pass,
        public readonly bool $matched,
    ) {
    }
}
