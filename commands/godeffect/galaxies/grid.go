package galaxies

import (
	"github.com/go-gl/mathgl/mgl64"
)

// Grid -
type Grid struct {
	Size    float64
	Spacing float64
}

// Generate - generate sphere
func (g Grid) Generate() (stars Stars) {
	count := int(g.Size / g.Spacing)

	for i := 0; i < count; i++ {
		for j := 0; j < count; j++ {
			for k := 0; k < count; k++ {
				pos := mgl64.Vec3{
					float64(i) * g.Spacing,
					float64(j) * g.Spacing,
					float64(k) * g.Spacing,
				}

				star := Star{Name: "TestName", Size: g.Size, Position: pos}

				star.Offset(mgl64.Vec3{
					-g.Size / 2,
					-g.Size / 2,
					-g.Size / 2,
				})

				stars = append(stars, star)
			}
		}
	}

	return
}
