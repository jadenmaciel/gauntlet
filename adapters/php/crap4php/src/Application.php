<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class Application
{
    /**
     * @param list<string> $args
     * @param resource|null $stdout
     * @param resource|null $stderr
     */
    public function run(array $args, $stdout = null, $stderr = null): int
    {
        $stdout = $stdout ?? STDOUT;
        $stderr = $stderr ?? STDERR;

        try {
            $options = (new OptionParser())->parse($args);
        } catch (HelpRequested) {
            fwrite($stdout, $this->usage());
            return 0;
        } catch (InvalidOption $error) {
            fwrite($stderr, "crap4php: {$error->getMessage()}\n");
            fwrite($stderr, $this->usage());
            return 2;
        }

        try {
            return $this->execute($options, $stdout, $stderr);
        } catch (\Throwable $error) {
            fwrite($stderr, "crap4php: {$error->getMessage()}\n");
            return 2;
        }
    }

    /**
     * @param resource $stdout
     * @param resource $stderr
     */
    private function execute(Options $options, $stdout, $stderr): int
    {
        $scanDir = realpath($options->dir);
        if ($scanDir === false) {
            throw new \RuntimeException(sprintf('scan directory not found: %s', $options->dir));
        }

        $normalizer = new PathNormalizer($scanDir);
        $scanner = new ComplexityScanner($normalizer);
        $functions = $scanner->scanDirectory($scanDir);

        $coverageByKey = [];
        if ($options->coverage !== null) {
            $coveragePath = $this->resolvePath($options->coverage);
            $coverageByKey = (new CloverCoverageParser($normalizer))
                ->coverageForFunctions($coveragePath, $functions);
        }

        $ceiling = $options->ceiling;
        if ($ceiling === null) {
            $thresholdPath = $this->resolvePath($options->thresholds);
            $ceiling = Thresholds::readCrapCeiling($thresholdPath);
        }

        $report = (new Evaluator())->evaluate($functions, $coverageByKey, $ceiling);
        (new ReportPrinter())->print($report, $options->format, $stdout);

        return $report->summary->failing > 0 ? 1 : 0;
    }

    private function resolvePath(string $path): string
    {
        if (str_starts_with($path, '/')) {
            return $path;
        }

        return getcwd() . DIRECTORY_SEPARATOR . $path;
    }

    private function usage(): string
    {
        return <<<TXT
Usage: crap4php [options]
  --dir <path>         directory to scan for .php files (default: .)
  --coverage <path>    path to PHPUnit clover XML coverage
  --ceiling <number>   override CRAP ceiling
  --thresholds <path>  thresholds file (default: .gauntlet/thresholds.yml)
  --format <text|json> output format (default: text)

TXT;
    }
}
