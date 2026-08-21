<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

final class Git
{
    public function __construct(private readonly string $dir)
    {
    }

    /**
     * Files touched since the merge base of $ref and HEAD, plus untracked ones.
     *
     * @return list<string>
     */
    public function changedFiles(string $ref): array
    {
        $base = trim($this->run(['merge-base', $ref, 'HEAD']));
        $diff = $this->run(['diff', '--name-only', $base]);
        $untracked = $this->run(['ls-files', '--others', '--exclude-standard']);

        $files = [];
        foreach (preg_split('/\R/', $diff . "\n" . $untracked) ?: [] as $line) {
            $line = trim($line);
            if ($line !== '') {
                $files[] = $line;
            }
        }

        return $files;
    }

    /**
     * @param list<string> $args
     */
    private function run(array $args): string
    {
        $command = 'git ' . implode(' ', array_map('escapeshellarg', $args));
        $descriptors = [1 => ['pipe', 'w'], 2 => ['pipe', 'w']];
        $process = proc_open($command, $descriptors, $pipes, $this->dir);
        if (!is_resource($process)) {
            throw new \RuntimeException(sprintf('could not run %s', $command));
        }

        $stdout = (string) stream_get_contents($pipes[1]);
        $stderr = (string) stream_get_contents($pipes[2]);
        fclose($pipes[1]);
        fclose($pipes[2]);

        if (proc_close($process) !== 0) {
            throw new \RuntimeException(sprintf('git %s: %s', implode(' ', $args), trim($stderr)));
        }

        return $stdout;
    }
}
