<?php

declare(strict_types=1);

function hotspot(int $n): int
{
    $sum = 0;
    if ($n > 0) { $sum++; }
    if ($n > 1) { $sum++; }
    if ($n > 2) { $sum++; }
    if ($n > 3) { $sum++; }
    if ($n > 4) { $sum++; }
    if ($n > 5) { $sum++; }
    return $sum;
}
