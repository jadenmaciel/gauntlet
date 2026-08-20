<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class JoinKey
{
    public static function fromParts(string $file, int $line, string $func): string
    {
        return sprintf('%s:%d:%s', $file, $line, $func);
    }

    public static function fromFunction(FunctionComplexity $function): string
    {
        return self::fromParts($function->file, $function->line, $function->func);
    }
}
