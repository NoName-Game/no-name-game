package godeffect

import (
	"log"

	"bitbucket.org/no-name-game/no-name/commands/godeffect/galaxies"
)

// Run -
func Run() (stars galaxies.Stars) {
	log.Println("Start")

	return generateSphere()
	// return generateGrid()
	// return generateCenter()
}

func generateGrid() (stars galaxies.Stars) {
	grid := galaxies.Grid{
		Size:    100.0,
		Spacing: 10.0,
	}

	return grid.Generate()
}

func generateSphere() galaxies.Stars {
	var stars galaxies.Stars

	for _, star := range singleShpere() {
		stars = append(stars, star)
	}

	for _, star := range secondShpere() {
		stars = append(stars, star)
	}

	return stars
}

func secondShpere() galaxies.Stars {
	var stars galaxies.Stars

	size := 20.0
	centerClusterDensityMean := 10.0
	centerClusterDensityDeviation := 0.0

	sphere := galaxies.Sphere{
		Size:             size,
		DensityMean:      centerClusterDensityMean,
		DensityDeviation: centerClusterDensityDeviation,
		DeviationX:       20,
		DeviationY:       20,
		DeviationZ:       20,
	}

	for _, star := range sphere.Generate() {
		stars = append(stars, star)
	}

	return stars
}

func singleShpere() galaxies.Stars {
	var stars galaxies.Stars

	size := 20.0
	centerClusterDensityMean := 10.0
	centerClusterDensityDeviation := 0.0

	sphere := galaxies.Sphere{
		Size:             size,
		DensityMean:      centerClusterDensityMean,
		DensityDeviation: centerClusterDensityDeviation,
		DeviationX:       1,
		DeviationY:       1,
		DeviationZ:       1,
	}

	for _, star := range sphere.Generate() {
		stars = append(stars, star)
	}

	return stars
}

// func generateCenter() (stars galaxies.Stars) {
// 	size := 1000.0                            // 750.0
// 	centerClusterScale := 0.19                // 0.19
// 	centerClusterDensityMean := 0.00005       // 0.00005
// 	centerClusterDensityDeviation := 0.000005 // 0.000005
// 	// centerClusterSizeDeviation := 0.00125

// 	centerClusterCountMean := 20.0          // 20.0
// 	centerClusterCountDeviation := 3.0      // 3.0
// 	centerClusterPositionDeviation := 0.075 //0.075

// 	swilr := (math.Pi * 4) * 5

// 	sphere := galaxies.Sphere{
// 		Size:             size * centerClusterScale,
// 		DensityMean:      centerClusterDensityMean,
// 		DensityDeviation: centerClusterDensityDeviation,
// 		DeviationX:       centerClusterScale,
// 		DeviationY:       centerClusterScale,
// 		DeviationZ:       centerClusterScale,
// 	}

// 	cluster := galaxies.Cluster{
// 		Basis:          sphere,
// 		CountMean:      centerClusterCountMean,
// 		CountDeviation: centerClusterCountDeviation,
// 		DeviationX:     size * centerClusterPositionDeviation,
// 		DeviationY:     size * centerClusterPositionDeviation,
// 		DeviationZ:     size * centerClusterPositionDeviation,
// 	}

// 	for _, star := range cluster.Generate() {
// 		star.Swirl(mgl64.Vec3{0, 1, 0}, swilr)
// 		stars = append(stars, star)
// 	}

// 	return stars
// }
