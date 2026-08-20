<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class FunctionComplexity
{
    public function __construct(
        public readonly string $file,
        public readonly int $line,
        public readonly int $endLine,
        public readonly string $func,
        public readonly int $complexity,
    ) {
    }
}
