<?php
namespace Goders\Galaxies;

abstract class AbstractGalaxyStruct
{
    abstract public function generate();

    // public function normallyDistributedSingle($n1, $n2)
    // {
    //     return 0.2;
    // }

    public function normallyDistributed($n1, $n2, $n3, $n4)
    {
        return 0.2;

        // static float Normall yDistributedSingle(this Random random, float standardDeviation, float mean, float min, float max) {
        // var nMax = (max - mean) / standardDeviation;
        // var nMin = (min - mean) / standardDeviation;
        // var nRange = nMax - nMin;
        // var nMaxSq = nMax * nMax;
        // var nMinSq = nMin * nMin;
        // var subFrom = nMinSq;
        // if (nMin < 0 && 0 < nMax) subFrom = 0;
        // else if (nMax < 0) subFrom = nMaxSq;

        // var sigma = 0.0;
        // double u;
        // float z;
        // do  {
        //     z = nRange * (float)random.NextDouble() + nMin; // uniform[no r mMin, normMax]
        //     sigma = Math.Exp((subFr o m - z * z) / 2);
        //     u = random.NextDouble();
        // } while (u > sigma);

        // return z * standardDeviation + mean;
        // }
    }

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
