package galaxies

import (
	"math"
)

// ClusterGenerator - ClusterGenerator
func ClusterGenerator(countMean float64, countDeviation float64, deviationX float64, deviationY float64, deviationZ float64) {


	count := math.Max(0, NormallyDistributedSingle(countDeviation, countMean));
if count > 0 {
	for i := 0; i < count; i++ {
		center := []float64{
			NormallyDistributedSingle(deviationX, 0),
			NormallyDistributedSingle(deviationY, 0),
			NormallyDistributedSingle(deviationZ, 0),
		}


	}
}

            for (int i = 0; i < count; i++)
            {
                Vector3 center = new Vector3(
                    random.NormallyDistributedSingle(_deviationX, 0),
                    random.NormallyDistributedSingle(_deviationY, 0),
                    random.NormallyDistributedSingle(_deviationZ, 0)
                );

                foreach (var star in _basis.Generate(random))
                    yield return star.Offset(center);
            }

}
