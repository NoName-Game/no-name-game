<?php
namespace Goders;

class Star
{
    /** @var array */
    public $position;

    /** @var string */
    protected $name;

    /** @var float */
    public $size;

    /** @var float */
    public $temperature;

    public function __construct(array $position, string $name, float $temp = 0)
    {
        $this->position = $position;
        $this->name = $name;
        $this->temperature = $temp;
    }

    public function offset(array $offset)
    {
        foreach ($offset as $key => $value) {
            $this->position[$key] = $this->position[$key] + $value;
        }

        return $this;
    }
}
