<?php
namespace Goders;

require __DIR__ . '/../vendor/autoload.php';

use Goders\Galaxy;
use Goders\Galaxies\Sphere;
use Goders\Galaxies\Cluster;
use Goders\Galaxies\Spiral;

// $galaxyStructType = new Sphere(1.0);

// $galaxy = new Galaxy();
// return $galaxy->generate($galaxyStructType);


// $galaxy = new Galaxy();
// return $galaxy->generate(new Cluster(new Sphere(1.0)));

$galaxy = new Galaxy();
return $galaxy->generate(new Spiral());
