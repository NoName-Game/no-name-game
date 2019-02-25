<?php
namespace Goders;

use Goders\Galaxies\AbstractGalaxyStruct;

class Galaxy
{
    public static function generate(AbstractGalaxyStruct $galaxyStructType)
    {
        return $galaxyStructType->generate();
    }
}
