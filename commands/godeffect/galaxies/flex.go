package galaxies

import (
	"github.com/go-gl/mathgl/mgl64"
)

// Flex -
type Flex struct {
	Size    float64
	Spacing float64
}

// Generate - generate sphere
func (f Flex) Generate() (stars Stars) {
	count := int(f.Size / f.Spacing)

	for i := 0; i < count; i++ {
		for j := 0; j < count; j++ {
			for k := 0; k < count; k++ {

				pos := mgl64.Vec3{
					NormallyDistributedSingle(float64(i)*f.Spacing, 0),
					NormallyDistributedSingle(float64(j)*f.Spacing, 0),
					NormallyDistributedSingle(float64(k)*f.Spacing, 0),
				}

				// NormallyDistributedSingle(s.DeviationX*s.Size, 0),
				// pos := mgl64.Vec3{
				// 	float64(i) * f.Spacing,
				// 	float64(j) * f.Spacing,
				// 	float64(k) * f.Spacing,
				// }

				star := Star{Name: "TestName", Size: f.Size, Position: pos}

				// star.Offset(mgl64.Vec3{
				// 	-f.Size / 2,
				// 	-f.Size / 2,
				// 	-f.Size / 2,
				// })

				stars = append(stars, star)
			}
		}
	}

	return
}
