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
        $xml = simplexml_load_file($cloverPath);
        if ($xml === false) {
            throw new \RuntimeException(sprintf('failed reading clover XML: %s', $cloverPath));
        }

        $coverageByFile = $this->loadFileCoverage($xml);
        $coverageByKey = [];

        foreach ($functions as $function) {
            $fileCoverage = $coverageByFile[$function->file] ?? null;
            if ($fileCoverage === null) {
                continue;
            }

            [$totalStatements, $coveredStatements] = $this->statementCoverage(
                $fileCoverage['statements'],
                $function->line,
                $function->endLine,
            );

            if ($totalStatements > 0) {
                $coverageByKey[JoinKey::fromFunction($function)] = ($coveredStatements / $totalStatements) * 100.0;
                continue;
            }

            $methodCount = $fileCoverage['methods'][$function->line] ?? null;
            if ($methodCount !== null) {
                $coverageByKey[JoinKey::fromFunction($function)] = $methodCount > 0 ? 100.0 : 0.0;
            }
        }

        return $coverageByKey;
    }

    /**
     * @return array<string,array{statements: array<int,int>, methods: array<int,int>}>
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
                    $methodCounts[$line] = $count;
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
     * @return array{int,int}
     */
    private function statementCoverage(array $statementCounts, int $startLine, int $endLine): array
    {
        $total = 0;
        $covered = 0;
        foreach ($statementCounts as $line => $count) {
            if ($line < $startLine || $line > $endLine) {
                continue;
            }
            $total++;
            if ($count > 0) {
                $covered++;
            }
        }

        return [$total, $covered];
    }
}
