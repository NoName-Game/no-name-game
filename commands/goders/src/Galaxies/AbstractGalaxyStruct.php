<?php
namespace Goders\Galaxies;

abstract class AbstractGalaxyStruct
{
    abstract public function generate();

    // https://arxiv.org/pdf/0907.4010.pdf
    public function normallyDistributed($standardDeviation, $mean, $min, $max)
    {
        $nMax = ($max - $mean) / $standardDeviation;
        $nMin = ($min - $mean) / $standardDeviation;
        $nRange = $nMax - $nMin;
        $nMaxSq = $nMax * $nMax;
        $nMinSq = $nMin * $nMin;
        $subFrom = $nMinSq;

        if ($nMin < 0 && 0 < $nMax) {
            $subFrom = 0;
        } elseif ($nMax < 0) {
            $subFrom = $nMaxSq;
        }

        $sigma = 0.0;

        do {
            $z = $nRange * $this->randomFloat() + $nMin; // uniform[no r mMin, normMax]
            $sigma = exp(($subFrom - $z * $z) / 2);
            $u = $this->randomFloat();
        } while ($u > $sigma);

        return $z * $standardDeviation + $mean;
    }

    // https://en.wikipedia.org/wiki/Box%E2%80%93Muller_transform - Normal distriution single
    public function normallyDistributedSingle($standardDeviation, $mean)
    {
        $x = $this->randomFloat();
        $y = $this->randomFloat();

        return sqrt(-2 * log($x)) * cos(2 * pi() * $y) * $standardDeviation + $mean;
    }

    function randomFloat($min = 0, $max = 1)
    {
        return $min + mt_rand() / mt_getrandmax() * ($max - $min);
    }
}
