package godeffect

import (
	"log"
	"math"

	"bitbucket.org/no-name-game/no-name/commands/godeffect/galaxies"
	"github.com/go-gl/mathgl/mgl64"
)

// Run -
func Run() {
	log.Println("Start")

	generateCenter()
}

func generateCenter() {
	var stars galaxies.Stars

	size := 750.0
	centerClusterScale := 0.19
	centerClusterDensityMean := 0.00005
	centerClusterDensityDeviation := 0.000005
	// centerClusterSizeDeviation := 0.00125

	centerClusterCountMean := 20.0
	centerClusterCountDeviation := 3.0
	centerClusterPositionDeviation := 0.075

	swilr := math.Pi * 4

	sphere := galaxies.Sphere{
		Size:             size * centerClusterScale,
		DensityMean:      centerClusterDensityMean,
		DensityDeviation: centerClusterDensityDeviation,
		DeviationX:       centerClusterScale,
		DeviationY:       centerClusterScale,
		DeviationZ:       centerClusterScale,
	}

	cluster := galaxies.Cluster{
		Basis:          sphere,
		CountMean:      centerClusterCountMean,
		CountDeviation: centerClusterCountDeviation,
		DeviationX:     size * centerClusterPositionDeviation,
		DeviationY:     size * centerClusterPositionDeviation,
		DeviationZ:     size * centerClusterPositionDeviation,
	}

	for _, star := range cluster.Generate() {
		star.Swirl(mgl64.Vec3{0, 1, 0}, swilr*5)
		stars = append(stars, star)
	}

	log.Println(stars)
}
