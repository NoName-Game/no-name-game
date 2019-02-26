package galaxies

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl64"
)

// Sphere - Sphere
type Sphere struct {
	Size             float64
	DensityMean      float64
	DensityDeviation float64
	DeviationX       float64
	DeviationY       float64
	DeviationZ       float64
}

// Generate - generate sphere
func (s Sphere) Generate() (stars Stars) {
	density := math.Max(0, NormallyDistributedSingle(s.DensityDeviation, s.DensityMean))
	countMax := math.Max(0, (s.Size * s.Size * s.Size * density))

	if countMax > 0 {
		count := rand.Intn(int(countMax))
		for i := 0; i < count; i++ {
			pos := mgl64.Vec3{
				NormallyDistributedSingle(s.DeviationX*s.Size, 0),
				NormallyDistributedSingle(s.DeviationY*s.Size, 0),
				NormallyDistributedSingle(s.DeviationZ*s.Size, 0),
			}

			dimension := pos.Len() / s.Size
			mass := dimension*2000 + (1-dimension)*1500
			temperature := Distribution(4000, mass, 1000, 40000)
			stars = append(stars, Star{Name: "TestName", Size: s.Size, Position: pos, Temperature: temperature})
		}
	}

	return
}
