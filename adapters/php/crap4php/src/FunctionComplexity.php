<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class FunctionComplexity
{
    /**
     * @param list<array{int,int}> $nestedFunctionRanges
     */
    public function __construct(
        public readonly string $file,
        public readonly int $line,
        public readonly int $endLine,
        public readonly string $func,
        public readonly int $complexity,
        public readonly array $nestedFunctionRanges = [],
    ) {
    }
}
