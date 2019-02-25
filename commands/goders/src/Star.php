<?php
namespace Goders;

use pocketmine\math\Vector3;

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
}
