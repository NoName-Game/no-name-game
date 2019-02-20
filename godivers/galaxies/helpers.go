package galaxies

import (
	"math"
	"math/rand"
)

// NormallyDistributedSingle - NormallyDistributedSingle
func NormallyDistributedSingle(standardDeviation float64, mean float64) float64 {

	// for true {
	u1 := rand.Float64()
	u2 := rand.Float64()

	x1 := math.Sqrt(-2.0 * math.Log(u1))
	x2 := 2.0 * math.Pi * u2
	z1 := x1 * math.Sin(x2)

	return z1 * standardDeviation * mean
	// }
}

// NormallyDistributedMultiple - NormallyDistributedMultiple
func NormallyDistributedMultiple(standardDeviation float64, mean float64) (float64, float64) {

	// for true {
	u1 := rand.Float64()
	u2 := rand.Float64()

	x1 := math.Sqrt(-2.0 * math.Log(u1))
	x2 := 2.0 * math.Pi * u2

	z1 := x1 * math.Sin(x2)
	z2 := x1 * math.Cos(x2)

	return z1 * standardDeviation * mean, z2*standardDeviation + mean
	// }
}
