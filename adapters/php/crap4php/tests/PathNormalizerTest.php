<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\PathNormalizer;
use PHPUnit\Framework\TestCase;

final class PathNormalizerTest extends TestCase
{
    public function testRepoRelativeCoveragePathMatchesSourceDirectory(): void
    {
        $project = sys_get_temp_dir() . '/crap4php-path-' . bin2hex(random_bytes(8));
        $sourceDir = $project . '/src';
        if (!mkdir($sourceDir, 0777, true) && !is_dir($sourceDir)) {
            self::fail('failed creating source directory');
        }
        file_put_contents($sourceDir . '/Example.php', "<?php\n");
        $previousDirectory = getcwd();
        chdir($project);

        try {
            $normalizer = new PathNormalizer($sourceDir);

            $this->assertSame('Example.php', $normalizer->normalize('src/Example.php'));
            $this->assertSame('Example.php', $normalizer->normalize('Example.php'));
        } finally {
            if ($previousDirectory !== false) {
                chdir($previousDirectory);
            }
        }
    }
}
