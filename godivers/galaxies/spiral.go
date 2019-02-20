package galaxies

import (
	"log"
	"math"
	"math/rand"
)

var (
	size    = 750
	spacing = 5

	minimumArms = 3
	maximumArms = 7

	clusterCountDeviation  float64 = 0.35
	clusterCenterDeviation         = 0.2

	minArmClusterScale       = 0.02
	armClusterScaleDeviation = 0.02
	maxArmClusterScale       = 0.1

	centerClusterScale            = 0.19
	centerClusterDensityMean      = 0.00005
	centerClusterDensityDeviation = 0.000005
	centerClusterSizeDeviation    = 0.00125

	centerClusterCountMean         float64 = 20
	centerClusterCountDeviation    float64 = 3
	centerClusterPositionDeviation         = 0.075

	centralVoidSizeMean      float64 = 25
	centralVoidSizeDeviation float64 = 7
)

// Spiral - Generate spiral
func Spiral() {
	log.Panicln("hello")

	centralVoidSize := NormallyDistributedSingle(centralVoidSizeDeviation, centralVoidSizeMean)
	if centralVoidSize < 0 {
		centralVoidSize = 0
	}
	centralVoidSizeSqr := centralVoidSize * centralVoidSize

	// foreach (var star in GenerateArms(random))
	// if (star.Position.LengthSquared() > centralVoidSizeSqr)
	//     yield return star;

	log.Println(centralVoidSize)
}

func generateArms() {
	var arms int
	arms = rand.Intn(maximumArms-minimumArms) + minimumArms

	var armAngle float64
	armAngle = (math.Pi * 2) / float64(arms)

	var maxClusters int
	maxClusters = (size / spacing) / arms

	for arm := 0; arm < arms; arm++ {

		clusters := NormallyDistributedSingle((float64(maxClusters) * clusterCountDeviation), float64(maxClusters))

		// var clusters int
		// clusters = int(math.Round(N1))

		for i := 0; i < int(math.Round(clusters)); i++ {

			//Angle from center of this arm
			var angle float64
			angle = NormallyDistributedSingle(0.5*armAngle*clusterCenterDeviation, 0) + armAngle*float64(arm)

			//Distance along this arm
			var dist float64
			dist = math.Abs(NormallyDistributedSingle(float64(size)*0.4, 0))

			//Center of the cluster
			//var center = Vector3.Transform(new Vector3(0, 0, dist), Quaternion.CreateFromAxisAngle(new Vector3(0, 1, 0), angle));

			//Size of the cluster
			clsScaleDev := armClusterScaleDeviation * float64(size)
			clsScaleMin := minArmClusterScale * float64(size)
			clsScaleMax := maxArmClusterScale * float64(size)
			//TODO: Normale con scale
			// cSize := NormallyDistributedSingle(clsScaleDev, (clsScaleMin*0.5 + clsScaleMax*0.5), clsScaleMin, clsScaleMax)

			// TODO fare
			// var stars = new Sphere(cSize, densityMean: 0.00025f, deviationX: 1, deviationY: 1, deviationZ: 1).Generate(random);
			// foreach (var star in stars)
			// 		yield return star.Offset(center).Swirl(Vector3.UnitY, Swirl);
		}
	}
}
