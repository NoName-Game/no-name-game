<?php
namespace Goders\Galaxies;

class Spiral extends AbstractGalaxyStruct
{

    /** @var int */
    public $size;

    /** @var int */
    public $spacing;

    /** @var int */
    public $minimumArms;
    public $maximumArms;

    /** @var float */
    public $clusterCountDeviation;
    public $clusterCenterDeviation;

    /** @var float */
    public $minArmClusterScale;
    public $armClusterScaleDeviation;
    public $maxArmClusterScale;

    /** @var float */
    public $swirl;

    /** @var float */
    public $centerClusterScale;
    public $centerClusterDensityMean;
    public $centerClusterDensityDeviation;
    public $centerClusterSizeDeviation;

    /** @var float */
    public $centerClusterPositionDeviation;
    public $centerClusterCountDeviation;
    public $centerClusterCountMean;

    /** @var float */
    public $centralVoidSizeMean;
    public $centralVoidSizeDeviation;

    public function __construct()
    {
        $this->size = 750;
        $this->spacing = 5;

        $this->minimumArms = 3;
        $this->maximumArms = 7;

        $this->clusterCountDeviation = 0.35;
        $this->clusterCenterDeviation = 0.2;

        $this->minArmClusterScale = 0.02;
        $this->armClusterScaleDeviation = 0.02;
        $this->maxArmClusterScale = 0.1;

        $this->swirl = pi() * 4;

        $this->centerClusterScale = 0.19;
        $this->centerClusterDensityMean = 0.00005;
        $this->centerClusterDensityDeviation = 0.000005;
        $this->centerClusterSizeDeviation = 0.00125;

        $this->centerClusterCountMean = 20;
        $this->centerClusterCountDeviation = 3;
        $this->centerClusterPositionDeviation = 0.075;

        $this->centralVoidSizeMean = 25;
        $this->centralVoidSizeDeviation = 7;
    }

    public function generate()
    {
        $centralVoidSize = $this->normallyDistributedSingle($this->centralVoidSizeDeviation, $this->centralVoidSizeMean);
        if ($centralVoidSize < 0) {
            $centralVoidSize = 0;
        }
        $centralVoidSizeSqr = $centralVoidSize * $centralVoidSize;

        $this->generateArms();
    }


    public function generateArms()
    {
        $stars = [];

        $arms = rand($this->minimumArms, $this->maximumArms);
        $armAngle = ((pi()   * 2) / $arms);

        $maxClusters = ($this->size / $this->spacing) / $arms;

        for ($arm = 0; $arm < $arms; $arm++) {
            $clusters = round($this->normallyDistributedSingle($maxClusters * $this->clusterCountDeviation, $maxClusters));

            for ($i = 0; $i < $clusters; $i++) {
                //Angle from center of this arm
                $angle = $this->normallyDistributedSingle(0 . 5 * $armAngle * $this->clusterCenterDeviation, 0) + $armAngle * $arm;

                //Distance along this arm
                $dist = abs($this->normallyDistributedSingle($this->size * 0.4, 0));

                //Center of the cluster
                // var center = Vector3.Transform(new Vector3(0, 0, dist), Quaternion.CreateFromAxisAngle(new Vector3(0, 1, 0), angle));

                //Size of the clust e r
                $clsScaleDev = $this->armClusterScaleDeviation * $this->size;
                $clsScaleMin = $this->minArmClusterScale * $this->size;
                $clsScaleMax = $this->maxArmClusterScale * $this->size;
                $cSize = $this->normallyDistributedSingle($clsScaleDev, $clsScaleMin * 0.5 + $clsScaleMax * 0.5, $clsScaleMin, $clsScaleMax);

                // var stars = new Sphere(cSize, densityMea n: 0.00025f, deviationX: 1,  deviationY: 1, deviationZ: 1).Generate(rand o m);
                $sphereStars = new Sphere((float)$cSize, 0.00025, 0.000001, 1.0, 1.0, 1.0);
                foreach ($sphereStars->generate() as $star) {
                    var_dump($star);
                    die;
                }
            }
        }
    }
}
