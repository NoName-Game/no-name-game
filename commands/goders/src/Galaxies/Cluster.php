<?php
namespace Goders\Galaxies;

use Goders\Galaxies\AbstractGalaxyStruct;

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

    /** @var array */
    protected $stars;

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
        $count = max(0, $this->normallyDistributedSingle($this->countDeviation, $this->countMean));
        if ($count <= 0) {
            return;
        }

        for ($i = 0; $i < $count; $i++) {
            $center = [
                $this->normallyDistributedSingle($this->deviationX * $this->size, 0),
                $this->normallyDistributedSingle($this->deviationY * $this->size, 0),
                $this->normallyDistributedSingle($this->deviationZ * $this->size, 0),
            ];

            foreach ($this->basis->generate() as $star) {
                $this->stars = $star->offset($center);
            }
        }
    }
}
