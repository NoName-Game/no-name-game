package galaxies

import (
	"math"
	"math/rand"
)

// NormallyDistributedSingle - NormallyDistributedSingle
func NormallyDistributedSingle(standardDeviation float64, mean float64) float64 {
	x := rand.Float64()
	y := rand.Float64()

	return math.Sqrt(-2*math.Log(x))*math.Cos(2*math.Pi*y)*standardDeviation + mean
}

// Distribution - Distribution
func Distribution(standardDeviation float64, mean float64, min float64, max float64) float64 {
	nMax := (max - mean) / standardDeviation
	nMin := (min - mean) / standardDeviation
	nRange := nMax - nMin
	nMaxSq := nMax * nMax
	nMinSq := nMin * nMin
	subFrom := nMinSq

	if nMin < 0 && 0 < nMax {
		subFrom = 0
	} else if nMax < 0 {
		subFrom = nMaxSq
	}

	sigma := 0.0

	var u, z float64

	for {
		z = nRange*rand.Float64() + nMin // uniform[normMin, normMax]
		sigma = math.Exp((subFrom - z*z) / 2)
		u = rand.Float64()
		if u > sigma {
			break
		}
	}

	return z*standardDeviation + mean
}
