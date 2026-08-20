<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class PathNormalizer
{
    private string $baseDir;

    public function __construct(string $baseDir)
    {
        $resolved = realpath($baseDir);
        if ($resolved === false) {
            throw new \RuntimeException(sprintf('scan directory %s does not exist', $baseDir));
        }
        $this->baseDir = self::normalizeSeparators($resolved);
    }

    public function normalize(string $path): string
    {
        $candidate = self::normalizeSeparators($path);
        if (!self::isAbsolute($candidate)) {
            $candidate = $this->baseDir . '/' . $candidate;
        }

        $resolved = realpath($candidate);
        if ($resolved !== false) {
            $candidate = self::normalizeSeparators($resolved);
        } else {
            $candidate = self::collapsePath($candidate);
        }

        $prefix = $this->baseDir . '/';
        if (str_starts_with($candidate, $prefix)) {
            return substr($candidate, strlen($prefix));
        }
        if ($candidate === $this->baseDir) {
            return '.';
        }

        return $candidate;
    }

    private static function isAbsolute(string $path): bool
    {
        if (str_starts_with($path, '/')) {
            return true;
        }

        return (bool) preg_match('/^[A-Za-z]:\//', $path);
    }

    private static function normalizeSeparators(string $path): string
    {
        return str_replace('\\', '/', $path);
    }

    private static function collapsePath(string $path): string
    {
        $path = self::normalizeSeparators($path);
        $prefix = '';
        if (str_starts_with($path, '/')) {
            $prefix = '/';
            $path = substr($path, 1);
        } elseif (preg_match('/^[A-Za-z]:\//', $path) === 1) {
            $prefix = substr($path, 0, 3);
            $path = substr($path, 3);
        }

        $parts = explode('/', $path);
        $stack = [];
        foreach ($parts as $part) {
            if ($part === '' || $part === '.') {
                continue;
            }
            if ($part === '..') {
                if ($stack !== [] && end($stack) !== '..') {
                    array_pop($stack);
                    continue;
                }
                if ($prefix === '') {
                    $stack[] = '..';
                }
                continue;
            }
            $stack[] = $part;
        }

        return $prefix . implode('/', $stack);
    }
}
