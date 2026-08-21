<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class CloverCoverageParser
{
    private PathNormalizer $normalizer;

    public function __construct(PathNormalizer $normalizer)
    {
        $this->normalizer = $normalizer;
    }

    /**
     * @param list<FunctionComplexity> $functions
     * @return array<string,float> Coverage keyed as file:line:func, value in percent 0..100.
     */
    public function coverageForFunctions(string $cloverPath, array $functions): array
    {
        // Checked before parsing: simplexml_load_file() raises a bare PHP I/O
        // warning for a missing path, which buries the actual problem.
        if (!is_file($cloverPath)) {
            throw new \RuntimeException(sprintf(
                'coverage report not found: %s (generate one with `phpunit --coverage-clover`)',
                $cloverPath
            ));
        }

        $xml = simplexml_load_file($cloverPath);
        if ($xml === false) {
            throw new \RuntimeException(sprintf('failed reading clover XML: %s', $cloverPath));
        }

        $coverageByFile = $this->loadFileCoverage($xml);
        $coverageByKey = [];
        $methodOffsets = [];

        foreach ($functions as $function) {
            $fileCoverage = $coverageByFile[$function->file] ?? null;
            if ($fileCoverage === null) {
                continue;
            }

            [$totalStatements, $coveredStatements] = $this->statementCoverage(
                $fileCoverage['statements'],
                $function->line,
                $function->endLine,
                $function->nestedFunctionRanges,
            );

            if ($totalStatements > 0) {
                $coverageByKey[JoinKey::fromFunction($function)] = ($coveredStatements / $totalStatements) * 100.0;
                continue;
            }

            $methodName = self::shortFunctionName($function->func);
            $methodCounts = $fileCoverage['methods'][$function->line][$methodName] ?? [];
            $offsetKey = $function->file . ':' . $function->line . ':' . $methodName;
            $offset = $methodOffsets[$offsetKey] ?? 0;
            $methodCount = $methodCounts[$offset] ?? null;
            if ($methodCount !== null) {
                $coverageByKey[JoinKey::fromFunction($function)] = $methodCount > 0 ? 100.0 : 0.0;
                $methodOffsets[$offsetKey] = $offset + 1;
            }
        }

        return $coverageByKey;
    }

    /**
     * @return array<string,array{statements: array<int,int>, methods: array<int,array<string,list<int>>>}>
     */
    private function loadFileCoverage(\SimpleXMLElement $xml): array
    {
        $result = [];
        /** @var \SimpleXMLElement[] $files */
        $files = $xml->xpath('//file') ?: [];
        foreach ($files as $fileNode) {
            $rawPath = (string) $fileNode['name'];
            if ($rawPath === '') {
                continue;
            }
            $file = $this->normalizer->normalize($rawPath);
            $statementCounts = [];
            $methodCounts = [];

            /** @var \SimpleXMLElement[] $lines */
            $lines = $fileNode->xpath('./line') ?: [];
            foreach ($lines as $lineNode) {
                $type = (string) $lineNode['type'];
                $line = (int) $lineNode['num'];
                $count = (int) $lineNode['count'];
                if ($line <= 0) {
                    continue;
                }
                if ($type === 'stmt') {
                    $statementCounts[$line] = $count;
                }
                if ($type === 'method') {
                    $name = (string) $lineNode['name'];
                    if ($name !== '') {
                        $methodCounts[$line][$name][] = $count;
                    }
                }
            }

            $result[$file] = [
                'statements' => $statementCounts,
                'methods' => $methodCounts,
            ];
        }

        return $result;
    }

    /**
     * @param array<int,int> $statementCounts
     * @param list<array{int,int}> $excludedRanges
     * @return array{int,int}
     */
    private function statementCoverage(
        array $statementCounts,
        int $startLine,
        int $endLine,
        array $excludedRanges,
    ): array {
        $total = 0;
        $covered = 0;
        foreach ($statementCounts as $line => $count) {
            if (
                $line < $startLine
                || $line > $endLine
                || self::lineInRanges($line, $excludedRanges)
            ) {
                continue;
            }
            $total++;
            if ($count > 0) {
                $covered++;
            }
        }

        return [$total, $covered];
    }

    private static function shortFunctionName(string $function): string
    {
        $methodSeparator = strrpos($function, '::');
        if ($methodSeparator !== false) {
            return substr($function, $methodSeparator + 2);
        }

        $namespaceSeparator = strrpos($function, '\\');
        if ($namespaceSeparator !== false) {
            return substr($function, $namespaceSeparator + 1);
        }

        return $function;
    }

    /**
     * @param list<array{int,int}> $ranges
     */
    private static function lineInRanges(int $line, array $ranges): bool
    {
        foreach ($ranges as [$startLine, $endLine]) {
            if ($line > $startLine && $line <= $endLine) {
                return true;
            }
        }

        return false;
    }
}
