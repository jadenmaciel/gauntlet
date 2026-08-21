<?php

declare(strict_types=1);

namespace Gauntlet\Crap4Php\Tests;

use Gauntlet\Crap4Php\ChangedFileFilter;
use Gauntlet\Crap4Php\FunctionComplexity;
use PHPUnit\Framework\TestCase;

final class ChangedFileFilterTest extends TestCase
{
    public function testMatchesAllowsAPathSegmentSuffixInEitherDirection(): void
    {
        $this->assertTrue(ChangedFileFilter::matches('src/Hot.php', 'src/Hot.php'));
        $this->assertTrue(ChangedFileFilter::matches('Hot.php', 'src/Hot.php'));
        $this->assertTrue(ChangedFileFilter::matches('src/Hot.php', 'Hot.php'));
        $this->assertTrue(ChangedFileFilter::matches('lib\\src\\Hot.php', 'src/Hot.php'));
    }

    public function testMatchesRequiresAWholeSegment(): void
    {
        $this->assertFalse(ChangedFileFilter::matches('NotHot.php', 'Hot.php'));
        $this->assertFalse(ChangedFileFilter::matches('src/A.php', 'src/B.php'));
    }

    public function testApplyKeepsOnlyFunctionsInChangedFiles(): void
    {
        $kept = new FunctionComplexity('src/Hot.php', 3, 9, 'hot', 4);
        $dropped = new FunctionComplexity('src/Cold.php', 3, 5, 'cold', 1);

        $result = ChangedFileFilter::apply([$kept, $dropped], ['src/Hot.php']);

        $this->assertSame([$kept], $result);
    }

    public function testApplyWithNoChangedFilesKeepsNothing(): void
    {
        $function = new FunctionComplexity('src/Hot.php', 3, 9, 'hot', 4);

        $this->assertSame([], ChangedFileFilter::apply([$function], []));
    }
}
