<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class ChangedFileFilter
{
    /**
     * @param list<FunctionComplexity> $functions
     * @param list<string> $changedFiles
     * @return list<FunctionComplexity>
     */
    public static function apply(array $functions, array $changedFiles): array
    {
        $kept = [];
        foreach ($functions as $function) {
            foreach ($changedFiles as $candidate) {
                if (self::matches($function->file, $candidate)) {
                    $kept[] = $function;
                    break;
                }
            }
        }

        return $kept;
    }

    /**
     * Compare two paths allowing a path-segment suffix match in either
     * direction, because `git diff --name-only` reports paths from the repo
     * root while scanned files are relative to --dir, which may sit below it.
     */
    public static function matches(string $left, string $right): bool
    {
        $left = str_replace('\\', '/', $left);
        $right = str_replace('\\', '/', $right);

        return $left === $right
            || str_ends_with($left, '/' . $right)
            || str_ends_with($right, '/' . $left);
    }
}
