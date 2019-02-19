package main

import (
	"log"
	"math"
	"math/rand"
)

var (
	starNames = []string{
		"Acamar", "Achernar", "Achird", "Acrab",
	}

	size    int = 750
	spacing int = 5

	minimumArms int = 3
	maximumArms int = 7

	clusterCountDeviation  float64 = 0.35
	clusterCenterDeviation float64 = 0.2

	minArmClusterScale       float64 = 0.02
	armClusterScaleDeviation float64 = 0.02
	maxArmClusterScale       float64 = 0.1

	swirl float64 = math.Pi * 4

	centerClusterScale            float64 = 0.19
	centerClusterDensityMean      float64 = 0.00005
	centerClusterDensityDeviation float64 = 0.000005
	centerClusterSizeDeviation    float64 = 0.00125

	centerClusterCountMean         float64 = 20
	centerClusterCountDeviation    float64 = 3
	centerClusterPositionDeviation float64 = 0.075

	centralVoidSizeMean      float64 = 25
	centralVoidSizeDeviation float64 = 7
)

func main() {
	spiral()
}

//Galaxy -
type Galaxy struct {
	Stars []Star
}

//Star -
type Star struct {
	Name        string
	Size        float32
	Position    []float32
	Temperature float32
}

// NormallyDistributedSingle - NormallyDistributedSingle
func NormallyDistributedSingle(standardDeviation float64, mean float64) float64 {

	// for true {
	u1 := rand.Float64()
	u2 := rand.Float64()

	x1 := math.Sqrt(-2.0 * math.Log(u1))
	x2 := 2.0 * math.Pi * u2
	z1 := x1 * math.Sin(x2)

	return z1 * standardDeviation * mean
	// }
}

func spiral() {
	log.Println("hello")

	centralVoidSize := NormallyDistributedSingle(centralVoidSizeDeviation, centralVoidSizeMean)
	if centralVoidSize < 0 {
		centralVoidSize = 0
	}
	centralVoidSizeSqr := centralVoidSize * centralVoidSize

	log.Println(centralVoidSize)
}

func generateArms() {
	var arms int
	arms = rand.Intn(maximumArms-minimumArms) + minimumArms

	var armAngle float64
	armAngle = ((math.Pi * 2) / arms)

	var maxClusters int
	maxClusters = (size / spacing) / arms

	for arm := 0; arm < arms; arm++ {
		var clusters int
		clusters = math.Round(NormallyDistributedSingle(maxClusters*clusterCountDeviation, maxClusters))

		for i := 0; i < clusters; i++ {

			//Angle from center of this arm
			var angle float64
			angle = NormallyDistributedSingle(0.5*armAngle*clusterCenterDeviation, 0) + armAngle*arm

			//Distance along this arm
			var dist float64
			dist = math.Abs(NormallyDistributedSingle(size*0.4, 0))

			//Center of the cluster
			//var center = Vector3.Transform(new Vector3(0, 0, dist), Quaternion.CreateFromAxisAngle(new Vector3(0, 1, 0), angle));

			//Size of the cluster
			clsScaleDev := armClusterScaleDeviation * size
			clsScaleMin := minArmClusterScale * size
			clsScaleMax := maxArmClusterScale * size
			//TODO: Normale con scale
			cSize := NormallyDistributedSingle(clsScaleDev, (clsScaleMin*0.5 + clsScaleMax*0.5), clsScaleMin, clsScaleMax)

			// TODO fare
			// var stars = new Sphere(cSize, densityMean: 0.00025f, deviationX: 1, deviationY: 1, deviationZ: 1).Generate(random);
			// foreach (var star in stars)
			// 		yield return star.Offset(center).Swirl(Vector3.UnitY, Swirl);
		}
	}
}

// protected internal override IEnumerable<Star> Generate(Random random)
// {
//     var centralVoidSize = random.NormallyDistributedSingle(CentralVoidSizeDeviation, CentralVoidSizeMean);
//     if (centralVoidSize < 0)
//         centralVoidSize = 0;
//     var centralVoidSizeSqr = centralVoidSize * centralVoidSize;

//     foreach (var star in GenerateArms(random))
//         if (star.Position.LengthSquared() > centralVoidSizeSqr)
//             yield return star;

//     foreach (var star in GenerateCenter(random))
//         if (star.Position.LengthSquared() > centralVoidSizeSqr)
//             yield return star;

//     foreach (var star in GenerateBackgroundStars(random))
//         if (star.Position.LengthSquared() > centralVoidSizeSqr)
//             yield return star;
// }
