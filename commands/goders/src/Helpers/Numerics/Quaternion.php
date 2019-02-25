<?php
namespace Goders\Helpers\Numerics;

class Quaternion
{
    /** @var float */
    public $x;

    /** @var float */
    public $y;

    /** @var float */
    public $z;

    /** @var float */
    public $w;

    // Constructs a Quaternion from the given vector and rotation parts.
    public function __construct(float $x, float $y, float $z, float $w)
    {
        $this->x = $x;
        $this->y = $y;
        $this->z = $z;
        $this->w = $w;
    }

    // https://referencesource.microsoft.com/#System.Numerics/System/Numerics/Quaternion.cs
    public static function createFromAxisAngle(Vector3 $axis, float $angle)
    {
        $ans = new Quaternion(0.0, 0.0, 0.0, 1.0);

        $halfAngle = $angle * 0.5;
        $s = (float)sin($halfAngle);
        $c = (float)cos($halfAngle);

        $ans->x = $axis->x * $s;
        $ans->y = $axis->y * $s;
        $ans->z = $axis->z * $s;
        $ans->w = $c;

        return $ans;
    }
}
