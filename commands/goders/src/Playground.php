<?php
namespace Goders;

require __DIR__ . '/../vendor/autoload.php';

use Goders\Galaxy;
use Goders\Galaxies\Spiral;

$stars = Galaxy::generate(new Spiral());

echo json_encode($stars);
die;

// var_dump($stars);
