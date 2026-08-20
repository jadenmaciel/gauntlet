<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\Evaluator;
use Gauntlet\Crap4Php\FunctionComplexity;
use PHPUnit\Framework\TestCase;

final class EvaluatorTest extends TestCase
{
    public function testUnmatchedFunctionDefaultsToZeroCoverageAndIsStillScored(): void
    {
        $functions = [
            new FunctionComplexity(
                file: 'src/Example.php',
                line: 10,
                endLine: 20,
                func: 'Example::hotspot',
                complexity: 5,
            ),
        ];

        $report = (new Evaluator())->evaluate($functions, [], 20.0);

        $this->assertCount(1, $report->functions);
        $this->assertSame(1, $report->summary->total);
        $this->assertFalse($report->functions[0]->matched);
        $this->assertSame(0.0, $report->functions[0]->coverage);
        $this->assertGreaterThan(0.0, $report->functions[0]->crap);
    }
}
