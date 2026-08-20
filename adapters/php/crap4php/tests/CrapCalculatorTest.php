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
        $contents = file_get_contents(__DIR__ . '/../../../../testdata/crap-golden.json');
        if ($contents === false) {
            throw new \RuntimeException('failed reading shared CRAP golden vectors');
        }
        $vectors = json_decode($contents, true, 512, JSON_THROW_ON_ERROR);
        $cases = [];
        foreach ($vectors as $vector) {
            $cases[$vector['name']] = [
                $vector['complexity'],
                (float) $vector['coverage'],
                (float) $vector['crap'],
            ];
        }

        return $cases;
    }

    #[DataProvider('goldenVectors')]
    public function testGoldenVector(int $complexity, float $coveragePercent, float $expected): void
    {
        $actual = CrapCalculator::compute($complexity, $coveragePercent);

        $this->assertEqualsWithDelta($expected, $actual, 1e-9);
    }
}
