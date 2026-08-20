<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php;

use PhpParser\Node;
use PhpParser\Node\Expr;
use PhpParser\Node\FunctionLike;
use PhpParser\Node\Stmt;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor;
use PhpParser\NodeVisitorAbstract;
use PhpParser\Parser;
use PhpParser\ParserFactory;

final class ComplexityScanner
{
    private Parser $parser;
    private PathNormalizer $normalizer;

    public function __construct(PathNormalizer $normalizer, ?Parser $parser = null)
    {
        $this->normalizer = $normalizer;
        $this->parser = $parser ?? (new ParserFactory())->createForNewestSupportedVersion();
    }

    /**
     * @return list<FunctionComplexity>
     */
    public function scanDirectory(string $dir): array
    {
        $iterator = new \RecursiveIteratorIterator(
            new \RecursiveDirectoryIterator($dir, \FilesystemIterator::SKIP_DOTS),
        );

        $functions = [];
        foreach ($iterator as $fileInfo) {
            if (!$fileInfo instanceof \SplFileInfo || !$fileInfo->isFile()) {
                continue;
            }

            $path = $fileInfo->getPathname();
            if (!$this->isPhpSource($path)) {
                continue;
            }
            if ($this->pathContainsSegment($path, 'vendor')) {
                continue;
            }

            $functions = array_merge($functions, $this->scanFile($path));
        }

        usort($functions, static function (FunctionComplexity $a, FunctionComplexity $b): int {
            if ($a->file !== $b->file) {
                return $a->file <=> $b->file;
            }
            if ($a->line !== $b->line) {
                return $a->line <=> $b->line;
            }

            return $a->func <=> $b->func;
        });

        return $functions;
    }

    /**
     * @return list<FunctionComplexity>
     */
    private function scanFile(string $path): array
    {
        $code = file_get_contents($path);
        if ($code === false) {
            throw new \RuntimeException(sprintf('failed reading %s', $path));
        }

        $ast = $this->parser->parse($code);
        if ($ast === null) {
            return [];
        }

        $traverser = new NodeTraverser();
        $collector = new FunctionCollector($this->normalizer->normalize($path));
        $traverser->addVisitor($collector);
        $traverser->traverse($ast);

        return $collector->functions();
    }

    private function isPhpSource(string $path): bool
    {
        return str_ends_with(strtolower($path), '.php');
    }

    private function pathContainsSegment(string $path, string $segment): bool
    {
        $needle = '/' . $segment . '/';
        $normalized = str_replace('\\', '/', $path) . '/';

        return str_contains($normalized, $needle);
    }
}

final class FunctionCollector extends NodeVisitorAbstract
{
    private string $file;
    private string $namespace = '';

    /** @var list<string> */
    private array $typeStack = [];

    /** @var list<FunctionComplexity> */
    private array $functions = [];

    public function __construct(string $file)
    {
        $this->file = $file;
    }

    public function enterNode(Node $node)
    {
        if ($node instanceof Stmt\Namespace_) {
            $this->namespace = $node->name?->toString() ?? '';
            return null;
        }

        if (
            $node instanceof Stmt\Class_
            || $node instanceof Stmt\Interface_
            || $node instanceof Stmt\Trait_
            || $node instanceof Stmt\Enum_
        ) {
            $name = $node->name?->toString();
            if ($name !== null) {
                $this->typeStack[] = $name;
            }
            return null;
        }

        if ($node instanceof Stmt\Function_) {
            $this->functions[] = $this->toFunctionComplexity(
                line: $node->getStartLine(),
                endLine: $node->getEndLine(),
                functionName: $this->qualifiedFunctionName($node->name->toString()),
                complexity: $this->complexityFor($node),
            );
            return null;
        }

        if ($node instanceof Stmt\ClassMethod) {
            if ($node->stmts === null) {
                return null;
            }
            $this->functions[] = $this->toFunctionComplexity(
                line: $node->getStartLine(),
                endLine: $node->getEndLine(),
                functionName: $this->qualifiedMethodName($node->name->toString()),
                complexity: $this->complexityFor($node),
            );
        }

        return null;
    }

    public function leaveNode(Node $node)
    {
        if (
            $node instanceof Stmt\Class_
            || $node instanceof Stmt\Interface_
            || $node instanceof Stmt\Trait_
            || $node instanceof Stmt\Enum_
        ) {
            $name = $node->name?->toString();
            if ($name !== null && $this->typeStack !== []) {
                array_pop($this->typeStack);
            }
        }

        return null;
    }

    /**
     * @return list<FunctionComplexity>
     */
    public function functions(): array
    {
        return $this->functions;
    }

    private function toFunctionComplexity(int $line, int $endLine, string $functionName, int $complexity): FunctionComplexity
    {
        return new FunctionComplexity(
            file: $this->file,
            line: $line,
            endLine: $endLine,
            func: $functionName,
            complexity: $complexity,
        );
    }

    private function qualifiedFunctionName(string $functionName): string
    {
        if ($this->namespace === '') {
            return $functionName;
        }

        return $this->namespace . '\\' . $functionName;
    }

    private function qualifiedMethodName(string $methodName): string
    {
        $type = $this->typeStack !== [] ? end($this->typeStack) : null;
        $prefixParts = [];
        if ($this->namespace !== '') {
            $prefixParts[] = $this->namespace;
        }
        if ($type !== null) {
            $prefixParts[] = $type;
        }

        if ($prefixParts === []) {
            return $methodName;
        }

        return implode('\\', $prefixParts) . '::' . $methodName;
    }

    private function complexityFor(FunctionLike $functionLike): int
    {
        $traverser = new NodeTraverser();
        $counter = new DecisionCounter();
        $traverser->addVisitor($counter);
        $traverser->traverse($functionLike->getStmts() ?? []);

        return $counter->complexity();
    }
}

final class DecisionCounter extends NodeVisitorAbstract
{
    private int $complexity = 1;

    public function enterNode(Node $node): ?int
    {
        if ($node instanceof FunctionLike) {
            return NodeVisitor::DONT_TRAVERSE_CHILDREN;
        }
        if (self::isDecisionNode($node)) {
            $this->complexity += self::decisionWeight($node);
        }

        return null;
    }

    public function complexity(): int
    {
        return $this->complexity;
    }

    private static function isDecisionNode(Node $node): bool
    {
        return $node instanceof Stmt\If_
            || $node instanceof Stmt\For_
            || $node instanceof Stmt\Foreach_
            || $node instanceof Stmt\While_
            || $node instanceof Stmt\Do_
            || $node instanceof Stmt\Catch_
            || $node instanceof Stmt\Case_
            || $node instanceof Expr\Ternary
            || $node instanceof Expr\BinaryOp\BooleanAnd
            || $node instanceof Expr\BinaryOp\BooleanOr
            || $node instanceof Expr\BinaryOp\LogicalAnd
            || $node instanceof Expr\BinaryOp\LogicalOr
            || $node instanceof Expr\BinaryOp\LogicalXor
            || $node instanceof Expr\Match_;
    }

    private static function decisionWeight(Node $node): int
    {
        if ($node instanceof Stmt\If_) {
            return 1 + count($node->elseifs);
        }
        if ($node instanceof Stmt\Case_) {
            return $node->cond === null ? 0 : 1;
        }
        if ($node instanceof Expr\Match_) {
            $count = 0;
            foreach ($node->arms as $arm) {
                if ($arm->conds === []) {
                    continue;
                }
                $count += count($arm->conds);
            }

            return $count;
        }

        return 1;
    }
}
