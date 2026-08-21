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
        $this->assertSame(100, $json['functions'][0]['coverage']);
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
     * @return array{int,string,string}
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
        rewind($stderr);
        $output = stream_get_contents($stdout);
        $errors = stream_get_contents($stderr);
        if ($output === false || $errors === false) {
            self::fail('failed reading output streams');
        }

        return [$exitCode, $output, $errors];
    }

    public function testChangedScopesTheReportToTouchedFiles(): void
    {
        $repo = $this->writeGitFixtureRepo(false);

        [$exitCode, $stdout] = $this->runApplication([
            '--dir', $repo,
            '--coverage', $repo . '/clover.xml',
            '--ceiling', '30',
            '--changed', 'HEAD',
            '--format', 'json',
        ]);

        $this->assertSame(0, $exitCode);
        $json = json_decode($stdout, true, 512, JSON_THROW_ON_ERROR);
        $files = array_values(array_unique(array_column($json['functions'], 'file')));
        $this->assertSame(['touched.php'], $files);
    }

    public function testChangedWithAnEmptyDiffWarnsInsteadOfPassingSilently(): void
    {
        $repo = $this->writeGitFixtureRepo(true);

        [$exitCode, $stdout, $stderr] = $this->runApplication([
            '--dir', $repo,
            '--coverage', $repo . '/clover.xml',
            '--ceiling', '1',
            '--changed', 'HEAD',
            '--format', 'json',
        ]);

        $this->assertSame(0, $exitCode);
        $this->assertStringContainsString('no files changed', $stderr);
        $json = json_decode($stdout, true, 512, JSON_THROW_ON_ERROR);
        $this->assertSame(0, $json['summary']['total']);
    }

    public function testChangedWithAnUnknownRefIsAUsageError(): void
    {
        $repo = $this->writeGitFixtureRepo(false);

        [$exitCode, , $stderr] = $this->runApplication([
            '--dir', $repo,
            '--coverage', $repo . '/clover.xml',
            '--ceiling', '30',
            '--changed', 'no-such-ref',
        ]);

        $this->assertSame(2, $exitCode);
        $this->assertStringContainsString('merge-base', $stderr);
    }

    public function testMissingCoverageFlagIsAUsageError(): void
    {
        [$exitCode, , $stderr] = $this->runApplication([
            '--dir', __DIR__ . '/fixtures',
            '--ceiling', '30',
        ]);

        $this->assertSame(2, $exitCode);
        $this->assertStringContainsString('--coverage is required', $stderr);
    }

    public function testMissingCoverageFileIsAUsageError(): void
    {
        $fixturesDir = __DIR__ . '/fixtures';

        [$exitCode] = $this->runApplication([
            '--dir', $fixturesDir,
            '--coverage', $fixturesDir . '/absent.xml',
            '--ceiling', '30',
        ]);

        $this->assertSame(2, $exitCode);
    }

    public function testMissingCeilingSourceIsAUsageError(): void
    {
        $fixturesDir = __DIR__ . '/fixtures';

        [$exitCode] = $this->runApplication([
            '--dir', $fixturesDir,
            '--coverage', $fixturesDir . '/clover-high.xml',
            '--thresholds', $fixturesDir . '/absent.yml',
        ]);

        $this->assertSame(2, $exitCode);
    }

    private function writeGitFixtureRepo(bool $commitEverything): string
    {
        $dir = sys_get_temp_dir() . '/crap4php-git-' . bin2hex(random_bytes(8));
        if (!mkdir($dir, 0777, true) && !is_dir($dir)) {
            self::fail('failed creating git fixture directory');
        }

        $body = "<?php\n\nfunction sample(int \$n): int\n{\n    if (\$n > 0) {\n        return 1;\n    }\n\n    return 0;\n}\n";
        file_put_contents($dir . '/committed.php', $body);
        file_put_contents(
            $dir . '/clover.xml',
            "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<coverage clover=\"3.2.0\"><project timestamp=\"0\"/></coverage>\n"
        );

        $this->git($dir, ['init', '-q']);
        $this->git($dir, ['add', '.']);
        $this->commit($dir, 'initial');

        file_put_contents($dir . '/touched.php', $body);
        if ($commitEverything) {
            $this->git($dir, ['add', '.']);
            $this->commit($dir, 'second');
        }

        return $dir;
    }

    /**
     * @param list<string> $args
     */
    private function commit(string $dir, string $message): void
    {
        $this->git($dir, [
            '-c', 'user.name=test',
            '-c', 'user.email=test@example.com',
            'commit', '-qm', $message,
        ]);
    }

    /**
     * @param list<string> $args
     */
    private function git(string $dir, array $args): void
    {
        // Ignore the developer's global and system git config. A global
        // core.hooksPath can drop generated files into the fixture on commit,
        // which then show up as untracked changes and skew the assertions.
        $command = 'GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null git '
            . implode(' ', array_map('escapeshellarg', $args))
            . ' 2>&1';
        $output = [];
        $status = 0;
        exec(sprintf('cd %s && %s', escapeshellarg($dir), $command), $output, $status);
        if ($status !== 0) {
            self::fail('git ' . implode(' ', $args) . ': ' . implode("\n", $output));
        }
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
