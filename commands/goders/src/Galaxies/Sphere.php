<?php
namespace Goders\Galaxies;

use Goders\Galaxies\AbstractGalaxyStruct;
use Goders\Star;
use Goders\Helpers\Numerics\Vector3;

class Sphere extends AbstractGalaxyStruct
{
    /** @var float */
    private $size;

    /** @var float */
    private $densityMean;

    /** @var float */
    private $densityDeviation;

    /** @var float */
    private $deviationX;
    private $deviationY;
    private $deviationZ;

    // /** @var array */
    // protected $stars;

    public function __construct(
        float $size,
        float $densityMean = 0.0000025,
        float $densityDeviation = 0.000001,
        float $deviationX = 0.0000025,
        float $deviationY = 0.0000025,
        float $deviationZ = 0.0000025
    ) {
        $this->size = $size;
        $this->densityMean = $densityMean;
        $this->densityDeviation = $densityDeviation;
        $this->deviationX = $deviationX;
        $this->deviationY = $deviationY;
        $this->deviationZ = $deviationZ;
    }

    public function generate()
    {
        $stars = [];
        $density = max(0, $this->normallyDistributedSingle($this->densityDeviation, $this->densityMean));
        $countMax = max(0, ($this->size * $this->size * $this->size * $density));

        if ($countMax > 0) {
            $count = mt_rand(0.1, $countMax);
            for ($i = 0; $i < $count; $i++) {

                $pos = new Vector3(
                    $this->normallyDistributedSingle($this->deviationX * $this->size, 0),
                    $this->normallyDistributedSingle($this->deviationY * $this->size, 0),
                    $this->normallyDistributedSingle($this->deviationZ * $this->size, 0)
                );

                $d = $pos->length() / $this->size;
                $m = $d * 2000 + (1 - $d) * 1500;
                $temperature = $this->normallyDistributed(4000, $m, 1000, 40000);

                $stars[] = new Star($pos, "Name-" . time(), (float)$temperature);
            }
        }

        return $stars;
    }
}
