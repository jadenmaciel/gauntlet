<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\CrapCalculator;
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

    public function testFunctionReportKeepsCoverageAsPercent(): void
    {
        $function = new FunctionComplexity(
            file: 'src/Example.php',
            line: 10,
            endLine: 20,
            func: 'Example::covered',
            complexity: 5,
        );
        $coverage = ['src/Example.php:10:Example::covered' => 75.0];

        $report = (new Evaluator())->evaluate([$function], $coverage, 30.0);

        $this->assertSame(75.0, $report->functions[0]->coverage);
        $this->assertSame(CrapCalculator::compute(5, 75.0), $report->functions[0]->crap);
    }
}
