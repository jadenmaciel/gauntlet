<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class Report
{
    /**
     * @param list<FunctionReport> $functions
     */
    public function __construct(
        public readonly float $ceiling,
        public readonly array $functions,
        public readonly Summary $summary,
    ) {
    }
}
