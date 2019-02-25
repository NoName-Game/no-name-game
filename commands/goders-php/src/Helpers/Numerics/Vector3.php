<?php
namespace Goders\Helpers\Numerics;

use pocketmine\math\Vector3 as BaseVector3;

class Vector3 extends BaseVector3
{
    public static function transform(Vector3 $value, Quaternion $rotation)
    {
        $x2 = $rotation->x + $rotation->x;
        $y2 = $rotation->y + $rotation->y;
        $z2 = $rotation->z + $rotation->z;

        $wx2 = $rotation->w * $x2;
        $wy2 = $rotation->w * $y2;
        $wz2 = $rotation->w * $z2;

        $xx2 = $rotation->x * $x2;
        $xy2 = $rotation->x * $y2;
        $xz2 = $rotation->x * $z2;

        $yy2 = $rotation->y * $y2;
        $yz2 = $rotation->y * $z2;
        $zz2 = $rotation->z * $z2;

        return new Vector3(
            $value->x * (1.0 - $yy2 - $zz2) + $value->y * ($xy2 - $wz2) + $value->z * ($xz2 + $wy2),
            $value->x * ($xy2 + $wz2) + $value->y * (1.0 - $xx2 - $zz2) + $value->z * ($yz2 - $wx2),
            $value->x * ($xz2 -  $wy2) + $value->y * ($yz2 +  $wx2) + $value->z * (1.0 - $xx2 - $yy2)
        );
    }
}
