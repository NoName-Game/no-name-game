<?php
namespace Goders\Galaxies;

abstract class AbstractGalaxyStruct
{
    abstract public function generate();

    public function normallyDistributedSingle($n1, $n2)
    {
        return 2;
    }

    public function normallyDistributed($n1, $n2, $n3, $n4)
    {
        return 2;
    }
}
