<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class ReportPrinter
{
    /**
     * @param resource $stream
     */
    public function print(Report $report, string $format, $stream): void
    {
        if ($format === 'json') {
            $this->printJson($report, $stream);
            return;
        }

        $this->printText($report, $stream);
    }

    /**
     * @param resource $stream
     */
    private function printJson(Report $report, $stream): void
    {
        $payload = [
            'ceiling' => $report->ceiling,
            'functions' => array_map(
                static fn (FunctionReport $function): array => [
                    'file' => $function->file,
                    'line' => $function->line,
                    'func' => $function->func,
                    'complexity' => $function->complexity,
                    'coverage' => $function->coverage,
                    'crap' => $function->crap,
                    'pass' => $function->pass,
                ],
                $report->functions,
            ),
            'summary' => [
                'total' => $report->summary->total,
                'failing' => $report->summary->failing,
                'max_crap' => $report->summary->maxCrap,
            ],
        ];

        fwrite($stream, json_encode($payload, JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR) . PHP_EOL);
    }

    /**
     * @param resource $stream
     */
    private function printText(Report $report, $stream): void
    {
        foreach ($report->functions as $function) {
            $status = $function->pass ? 'PASS' : 'FAIL';
            $note = $function->matched ? '' : ' (no coverage data)';
            fwrite(
                $stream,
                sprintf(
                    "%s:%d:%s\tcomplexity=%d\tcoverage=%.1f%%\tcrap=%.2f\t%s%s\n",
                    $function->file,
                    $function->line,
                    $function->func,
                    $function->complexity,
                    $function->coverage,
                    $function->crap,
                    $status,
                    $note,
                ),
            );
        }

        fwrite(
            $stream,
            sprintf(
                "\nceiling=%.2f total=%d failing=%d max_crap=%.2f\n",
                $report->ceiling,
                $report->summary->total,
                $report->summary->failing,
                $report->summary->maxCrap,
            ),
        );
    }
}
