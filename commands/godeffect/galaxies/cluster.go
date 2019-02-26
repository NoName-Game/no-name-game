package galaxies

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

// Cluster - Cluster
type Cluster struct {
	Basis          Sphere
	CountMean      float64
	CountDeviation float64
	DeviationX     float64
	DeviationY     float64
	DeviationZ     float64
}

// Generate - generate sphere
func (c Cluster) Generate() (stars Stars) {
	count := math.Max(0, NormallyDistributedSingle(c.CountDeviation, c.CountMean))
	if count > 0 {
		for i := 0; i < int(count); i++ {
			center := mgl64.Vec3{
				NormallyDistributedSingle(c.DeviationX, 0),
				NormallyDistributedSingle(c.DeviationY, 0),
				NormallyDistributedSingle(c.DeviationZ, 0),
			}

			for _, star := range c.Basis.Generate() {
				star.Offset(center)
				stars = append(stars, star)
			}
		}
	}

	return
}
