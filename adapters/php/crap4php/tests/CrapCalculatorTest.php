<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\CrapCalculator;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class CrapCalculatorTest extends TestCase
{
    public static function goldenVectors(): array
    {
        return [
            'cc1-cov0' => [1, 0.0, 2.0],
            'cc1-cov100' => [1, 100.0, 1.0],
            'cc2-cov0' => [2, 0.0, 6.0],
            'cc3-cov50' => [3, 50.0, 4.125],
            'cc5-cov0' => [5, 0.0, 30.0],
            'cc5-cov50' => [5, 50.0, 8.125],
            'cc5-cov100' => [5, 100.0, 5.0],
            'cc8-cov100' => [8, 100.0, 8.0],
            'cc10-cov0' => [10, 0.0, 110.0],
            'cc10-cov80' => [10, 80.0, 10.8],
            'cc10-cov100' => [10, 100.0, 10.0],
            'cc15-cov100' => [15, 100.0, 15.0],
            'cc20-cov90' => [20, 90.0, 20.4],
        ];
    }

    #[DataProvider('goldenVectors')]
    public function testGoldenVector(int $complexity, float $coveragePercent, float $expected): void
    {
        $actual = CrapCalculator::compute($complexity, $coveragePercent);

        $this->assertEqualsWithDelta($expected, $actual, 1e-9);
    }
}
