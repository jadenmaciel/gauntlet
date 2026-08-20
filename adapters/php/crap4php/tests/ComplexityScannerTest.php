<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\ComplexityScanner;
use Gauntlet\Crap4Php\PathNormalizer;
use PHPUnit\Framework\TestCase;

final class ComplexityScannerTest extends TestCase
{
    public function testNestedClosureDecisionsDoNotIncreaseOuterComplexity(): void
    {
        $dir = $this->writeSource(<<<'PHP'
<?php

function outer(): void
{
    $inner = function (): void {
        if (true) {
        }
    };
}
PHP);

        $functions = (new ComplexityScanner(new PathNormalizer($dir)))->scanDirectory($dir);

        $this->assertCount(1, $functions);
        $this->assertSame('outer', $functions[0]->func);
        $this->assertSame(1, $functions[0]->complexity);
    }

    public function testDeclarationsWithoutBodiesAreNotReported(): void
    {
        $dir = $this->writeSource(<<<'PHP'
<?php

interface Contract
{
    public function run(): void;
}

abstract class Base
{
    abstract public function stop(): void;
}

final class Concrete
{
    public function work(): void
    {
    }
}
PHP);

        $functions = (new ComplexityScanner(new PathNormalizer($dir)))->scanDirectory($dir);

        $this->assertCount(1, $functions);
        $this->assertSame('Concrete::work', $functions[0]->func);
    }

    public function testMatchDefaultArmDoesNotCrashOrIncreaseComplexity(): void
    {
        $dir = $this->writeSource(<<<'PHP'
<?php

function classify(int $value): string
{
    return match ($value) {
        1, 2 => 'small',
        default => 'large',
    };
}
PHP);

        $functions = (new ComplexityScanner(new PathNormalizer($dir)))->scanDirectory($dir);

        $this->assertCount(1, $functions);
        $this->assertSame(3, $functions[0]->complexity);
    }

    public function testAnonymousClassMethodsAreNotAttributedToEnclosingClass(): void
    {
        $dir = $this->writeSource(<<<'PHP'
<?php

final class Factory
{
    public function make(): object
    {
        return new class {
            public function run(): void
            {
            }
        };
    }
}
PHP);

        $functions = (new ComplexityScanner(new PathNormalizer($dir)))->scanDirectory($dir);

        $this->assertCount(1, $functions);
        $this->assertSame('Factory::make', $functions[0]->func);
    }

    private function writeSource(string $source): string
    {
        $dir = sys_get_temp_dir() . '/crap4php-scanner-' . bin2hex(random_bytes(8));
        if (!mkdir($dir, 0777, true) && !is_dir($dir)) {
            self::fail('failed creating source directory');
        }
        if (file_put_contents($dir . '/Example.php', $source) === false) {
            self::fail('failed writing PHP source');
        }

        return $dir;
    }
}
