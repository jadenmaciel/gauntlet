<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class Evaluator
{
    /**
     * @param list<FunctionComplexity> $functions
     * @param array<string,float> $coverageByKey Coverage as percent 0..100.
     */
    public function evaluate(array $functions, array $coverageByKey, float $ceiling): Report
    {
        $functionReports = [];
        $failing = 0;
        $maxCrap = 0.0;

        foreach ($functions as $function) {
            $key = JoinKey::fromFunction($function);
            $matched = array_key_exists($key, $coverageByKey);
            $coveragePercent = $matched ? $coverageByKey[$key] : 0.0;
            $crap = CrapCalculator::compute($function->complexity, $coveragePercent);
            $coverageFraction = $coveragePercent / 100.0;
            $pass = $crap <= $ceiling;

            if (!$pass) {
                $failing++;
            }
            if ($crap > $maxCrap) {
                $maxCrap = $crap;
            }

            $functionReports[] = new FunctionReport(
                file: $function->file,
                line: $function->line,
                func: $function->func,
                complexity: $function->complexity,
                coverage: $coverageFraction,
                crap: $crap,
                pass: $pass,
                matched: $matched,
            );
        }

        return new Report(
            ceiling: $ceiling,
            functions: $functionReports,
            summary: new Summary(total: count($functions), failing: $failing, maxCrap: $maxCrap),
        );
    }
}
