package galaxies

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	randplus "k8s.io/apimachinery/pkg/util/rand"
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
	// var stars Stars

	density := math.Max(0, NormallyDistributedSingle(s.DensityDeviation, s.DensityMean))
	countMax := math.Max(0, (s.Size * s.Size * s.Size * density))

	if countMax > 0 {
		count := randplus.Intn(int(countMax))
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

	// for _, star := range stars {
	// 	log.Println(star)
	// }

}

// for ($i = 0; $i < $count; $i++) {

//     $pos = new Vector3(
//         $this->normallyDistributedSingle($this->deviationX * $this->size, 0),
//         $this->normallyDistributedSingle($this->deviationY * $this->size, 0),
//         $this->normallyDistributedSingle($this->deviationZ * $this->size, 0)
//     );

//     $d = $pos->length() / $this->size;
//     $m = $d * 2000 + (1 - $d) * 1500;
//     $temperature = $this->normallyDistributed(4000, $m, 1000, 40000);

//     $stars[] = new Star($pos, "Name-" . time(), (float)$temperature);
// }
