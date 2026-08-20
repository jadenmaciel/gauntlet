<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\CloverCoverageParser;
use Gauntlet\Crap4Php\ComplexityScanner;
use Gauntlet\Crap4Php\FunctionComplexity;
use Gauntlet\Crap4Php\JoinKey;
use Gauntlet\Crap4Php\PathNormalizer;
use PHPUnit\Framework\TestCase;

final class CloverCoverageParserTest extends TestCase
{
    public function testNestedClosureStatementsDoNotChangeOuterCoverage(): void
    {
        $dir = $this->makeTempDirectory();
        $source = <<<'PHP'
<?php

function outer(): callable
{
    $outer = 1;
    return function (): int {
        $nested = 2;
        return $nested;
    };
}
PHP;
        file_put_contents($dir . '/Example.php', $source);
        $clover = <<<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <project>
    <file name="Example.php">
      <line num="3" type="method" name="outer" count="1"/>
      <line num="5" type="stmt" count="0"/>
      <line num="7" type="stmt" count="1"/>
      <line num="8" type="stmt" count="1"/>
    </file>
  </project>
</coverage>
XML;
        $cloverPath = $dir . '/clover.xml';
        file_put_contents($cloverPath, $clover);
        $normalizer = new PathNormalizer($dir);
        $functions = (new ComplexityScanner($normalizer))->scanDirectory($dir);

        $coverage = (new CloverCoverageParser($normalizer))->coverageForFunctions($cloverPath, $functions);

        $this->assertCount(1, $coverage);
        $this->assertSame(0.0, array_values($coverage)[0]);
    }

    public function testNestedClosureDeclarationLineRemainsOuterCoverage(): void
    {
        $dir = $this->makeTempDirectory();
        $source = <<<'PHP'
<?php

function outer(): callable
{
    return function (): int {
        return 1;
    };
}
PHP;
        file_put_contents($dir . '/Example.php', $source);
        $clover = <<<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <project>
    <file name="Example.php">
      <line num="5" type="stmt" count="1"/>
      <line num="6" type="stmt" count="0"/>
    </file>
  </project>
</coverage>
XML;
        $cloverPath = $dir . '/clover.xml';
        file_put_contents($cloverPath, $clover);
        $normalizer = new PathNormalizer($dir);
        $functions = (new ComplexityScanner($normalizer))->scanDirectory($dir);

        $coverage = (new CloverCoverageParser($normalizer))->coverageForFunctions($cloverPath, $functions);

        $this->assertCount(1, $coverage);
        $this->assertSame(100.0, $coverage[JoinKey::fromFunction($functions[0])]);
    }

    public function testMethodFallbackUsesTheMethodNameWhenLinesAreShared(): void
    {
        $dir = $this->makeTempDirectory();
        file_put_contents($dir . '/Example.php', "<?php\n");
        $clover = <<<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <project>
    <file name="Example.php">
      <line num="4" type="method" name="covered" count="1"/>
      <line num="4" type="method" name="missed" count="0"/>
    </file>
  </project>
</coverage>
XML;
        $cloverPath = $dir . '/clover.xml';
        file_put_contents($cloverPath, $clover);
        $normalizer = new PathNormalizer($dir);
        $covered = new FunctionComplexity('Example.php', 4, 4, 'Example::covered', 1);
        $missed = new FunctionComplexity('Example.php', 4, 4, 'Example::missed', 1);

        $coverage = (new CloverCoverageParser($normalizer))->coverageForFunctions(
            $cloverPath,
            [$covered, $missed],
        );

        $this->assertSame(100.0, $coverage[JoinKey::fromFunction($covered)]);
        $this->assertSame(0.0, $coverage[JoinKey::fromFunction($missed)]);
    }

    private function makeTempDirectory(): string
    {
        $dir = sys_get_temp_dir() . '/crap4php-clover-' . bin2hex(random_bytes(8));
        if (!mkdir($dir, 0777, true) && !is_dir($dir)) {
            self::fail('failed creating temporary directory');
        }

        return $dir;
    }
}
