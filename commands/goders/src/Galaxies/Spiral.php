<?php
namespace Goders\Galaxies;

use pocketmine\math\Vector3;

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
        $this->size = 750; //750
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
        $stars = [];
        $centralVoidSize = $this->normallyDistributedSingle($this->centralVoidSizeDeviation, $this->centralVoidSizeMean);
        if ($centralVoidSize < 0) {
            $centralVoidSize = 0;
        }
        $centralVoidSizeSqr = $centralVoidSize * $centralVoidSize;

        // foreach ($this->generateArms() as $star) {
        //if (star.Position.LengthSquared() > centralVoidSizeSqr)
        // yield return star;
        // };

        foreach ($this->generateCenter() as $star) {
            if ($star->position->lengthSquared() > $centralVoidSizeSqr) {
                $stars[] = $star;
            }
        }

        foreach ($this->generateBackground() as $star) {
            if ($star->position->lengthSquared() > $centralVoidSizeSqr) {
                $stars[] = $star;
            }
        }

        var_dump($stars);
        die;
    }

    public function generateBackground()
    {
        $sphere = new Sphere((float)$this->size, 0.000001, 0.0000001, 0.35, 0.125, 0.35);
        return $sphere->generate();
    }

    public function generateCenter()
    {
        $stars = [];
        $sphere = new Sphere(
            (float)($this->size * $this->centerClusterScale),
            $this->centerClusterDensityMean,
            $this->centerClusterDensityDeviation,
            $this->centerClusterScale,
            $this->centerClusterScale,
            $this->centerClusterScale
        );

        $cluster = new Cluster(
            $sphere,
            (float)$this->centerClusterCountMean,
            (float)$this->centerClusterCountDeviation,
            (float)($this->size * $this->centerClusterPositionDeviation),
            (float)($this->size * $this->centerClusterPositionDeviation),
            (float)($this->size * $this->centerClusterPositionDeviation)
        );

        foreach ($cluster->generate() as $star) {
            $stars[] = $star;
        }

        return $stars;
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
