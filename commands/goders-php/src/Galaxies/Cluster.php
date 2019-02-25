<?php
namespace Goders\Galaxies;

use Goders\Galaxies\AbstractGalaxyStruct;
use Goders\Helpers\Numerics\Vector3;

class Cluster extends AbstractGalaxyStruct
{
    /** @var AbstractGalaxyStruct */
    private $basis;

    /** @var float */
    private $countMean;

    /** @var float */
    private $countDeviation;

    /** @var float */
    private $deviationX;
    private $deviationY;
    private $deviationZ;

    public function __construct(
        AbstractGalaxyStruct $basis,
        float $countMean = 0.0000025,
        float $countDeviation = 0.000001,
        float $deviationX = 0.0000025,
        float $deviationY = 0.0000025,
        float $deviationZ = 0.0000025
    ) {
        $this->basis = $basis;
        $this->countMean = $countMean;
        $this->countDeviation = $countDeviation;
        $this->deviationX = $deviationX;
        $this->deviationY = $deviationY;
        $this->deviationZ = $deviationZ;
    }

    public function generate()
    {
        $stars = [];
        $count = max(0, $this->normallyDistributedSingle($this->countDeviation, $this->countMean));
        if ($count > 0) {
            for ($i = 0; $i < $count; $i++) {
                $center = new Vector3(
                    $this->normallyDistributedSingle($this->deviationX, 0),
                    $this->normallyDistributedSingle($this->deviationY, 0),
                    $this->normallyDistributedSingle($this->deviationZ, 0)
                );

                foreach ($this->basis->generate() as $star) {
                    $stars[] = $star->offset($center);
                }
            }
        }

        return $stars;
    }
}
