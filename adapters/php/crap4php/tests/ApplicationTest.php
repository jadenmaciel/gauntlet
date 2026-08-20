<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\Application;
use PHPUnit\Framework\TestCase;

final class ApplicationTest extends TestCase
{
    public function testJsonFormatMatchesContractShape(): void
    {
        $fixturesDir = __DIR__ . '/fixtures';
        [$exitCode, $stdout] = $this->runApplication([
            '--dir', $fixturesDir,
            '--coverage', $fixturesDir . '/clover-high.xml',
            '--ceiling', '30',
            '--format', 'json',
        ]);

        $this->assertSame(0, $exitCode);
        $json = json_decode($stdout, true, 512, JSON_THROW_ON_ERROR);
        $this->assertSame(['ceiling', 'functions', 'summary'], array_keys($json));
        $this->assertSame(['file', 'line', 'func', 'complexity', 'coverage', 'crap', 'pass'], array_keys($json['functions'][0]));
        $this->assertSame(100.0, $json['functions'][0]['coverage']);
        $this->assertSame(['total', 'failing', 'max_crap'], array_keys($json['summary']));
    }

    public function testReadsCeilingFromThresholdsFileAndExitsBasedOnFailures(): void
    {
        $fixturesDir = __DIR__ . '/fixtures';
        $thresholds = $this->writeTempThresholds(30.0);

        [$failingExit, $failingOut] = $this->runApplication([
            '--dir', $fixturesDir,
            '--coverage', $fixturesDir . '/clover-low.xml',
            '--thresholds', $thresholds,
            '--format', 'json',
        ]);
        $failingJson = json_decode($failingOut, true, 512, JSON_THROW_ON_ERROR);

        [$passingExit, $passingOut] = $this->runApplication([
            '--dir', $fixturesDir,
            '--coverage', $fixturesDir . '/clover-high.xml',
            '--thresholds', $thresholds,
            '--format', 'json',
        ]);
        $passingJson = json_decode($passingOut, true, 512, JSON_THROW_ON_ERROR);

        $this->assertSame(1, $failingExit);
        $this->assertSame(0, $passingExit);
        $this->assertSame(1, $failingJson['summary']['failing']);
        $this->assertSame(0, $passingJson['summary']['failing']);
    }

    /**
     * @param list<string> $args
     * @return array{int,string}
     */
    private function runApplication(array $args): array
    {
        $stdout = fopen('php://temp', 'r+');
        $stderr = fopen('php://temp', 'r+');
        if ($stdout === false || $stderr === false) {
            self::fail('failed to create temp streams');
        }

        $exitCode = (new Application())->run($args, $stdout, $stderr);
        rewind($stdout);
        $output = stream_get_contents($stdout);
        if ($output === false) {
            self::fail('failed reading stdout');
        }

        return [$exitCode, $output];
    }

    private function writeTempThresholds(float $value): string
    {
        $dir = sys_get_temp_dir() . '/crap4php-test-' . bin2hex(random_bytes(8));
        if (!mkdir($dir . '/.gauntlet', 0777, true) && !is_dir($dir . '/.gauntlet')) {
            self::fail('failed creating thresholds directory');
        }
        $path = $dir . '/.gauntlet/thresholds.yml';
        $yaml = <<<YAML
metrics:
  crap_ceiling:
    direction: max
    value: {$value}
YAML;
        if (file_put_contents($path, $yaml) === false) {
            self::fail('failed writing thresholds file');
        }

        return $path;
    }
}
