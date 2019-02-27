package godeffect

import (
	"log"
	"math"

	"bitbucket.org/no-name-game/no-name/commands/godeffect/galaxies"
	"github.com/go-gl/mathgl/mgl64"
)

// Run -
func Run() (stars galaxies.Stars) {
	log.Println("Start")

	return generateFlex()
	// return generateGrid()
	// return generateCenter()
}

func generateFlex() (stars galaxies.Stars) {
	flex := galaxies.Flex{
		Size:    100.0,
		Spacing: 10.0,
	}

	return flex.Generate()
}

func generateGrid() (stars galaxies.Stars) {
	grid := galaxies.Grid{
		Size:    100.0,
		Spacing: 10.0,
	}

	return grid.Generate()
}

func generateCenter() (stars galaxies.Stars) {
	size := 1000.0                            // 750.0
	centerClusterScale := 0.19                // 0.19
	centerClusterDensityMean := 0.00005       // 0.00005
	centerClusterDensityDeviation := 0.000005 // 0.000005
	// centerClusterSizeDeviation := 0.00125

	centerClusterCountMean := 20.0          // 20.0
	centerClusterCountDeviation := 3.0      // 3.0
	centerClusterPositionDeviation := 0.075 //0.075

	swilr := (math.Pi * 4) * 5

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
		star.Swirl(mgl64.Vec3{0, 1, 0}, swilr)
		stars = append(stars, star)
	}

	return stars
}
