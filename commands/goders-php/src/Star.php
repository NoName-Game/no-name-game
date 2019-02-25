<?php
namespace Goders;

use Goders\Helpers\Numerics\Vector3;
use Goders\Helpers\Numerics\Quaternion;

class Star
{
    /** @var Vector3 */
    public $position;

    /** @var string */
    protected $name;

    /** @var float */
    public $size;

    /** @var float */
    public $temperature;

    public function __construct(Vector3 $position, string $name, float $temp = 0)
    {
        $this->position = $position;
        $this->name = $name;
        $this->temperature = $temp;
    }

    public function offset(Vector3 $offset)
    {
        $this->position->add($offset);

        return $this;
    }

    public function swirl(Vector3 $axis, float $amount)
    {
        $d = $this->position->length();

        /** @var float $a*/
        $a = (float)pow($d, 0.1) * $amount;

        $this->position = Vector3::transform($this->position, Quaternion::createFromAxisAngle($axis, $a));

        return $this;
    }
}
