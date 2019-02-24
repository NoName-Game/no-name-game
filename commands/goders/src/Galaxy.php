<?php
namespace Goders;

use Goders\Galaxies\AbstractGalaxyStruct;

class Galaxy
{
    public function generate(AbstractGalaxyStruct $galaxyStructType)
    {
        $galaxyStructType->generate();
    }
}
