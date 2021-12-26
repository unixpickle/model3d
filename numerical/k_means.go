// Implementation of K-means++ adapted from smallpng:
// https://github.com/unixpickle/smallpng/blob/f492db016612a4e7f8f9f3735ff021d95752e0a7/smallpng/quantize.go

package numerical

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/unixpickle/essentials"
)

// KMeans stores the (intermediate) results of K-means
// clustering.
//
// This type can be created using NewKMeans() to create an
// initial clustering using K-means++. Then, Iterate() can
// be called repeatedly to refine the clustering as
// desired.
type KMeans struct {
	// Centers is the K-means cluster centers.
	Centers []Vec

	// Data stores all of the data points.
	Data []Vec
}

// NewKMeans creates a KMeans object with K-means++
// initialization.
func NewKMeans(data []Vec, numCenters int) *KMeans {
	if len(data) <= numCenters {
		return &KMeans{
			Centers: data,
			Data:    data,
		}
	}
	return &KMeans{
		Centers: kmeansPlusPlusInit(data, numCenters),
		Data:    data,
	}
}

// Iterate performs a step of k-means and returns the
// current MSE loss.
// If the MSE loss does not decrease, then the process has
// converged.
func (k *KMeans) Iterate() float64 {
	dim := len(k.Centers[0])
	centerSum := make([]Vec, len(k.Centers))
	for i := range centerSum {
		centerSum[i] = make(Vec, dim)
	}
	centerCount := make([]int, len(k.Centers))
	totalError := 0.0

	numProcs := runtime.GOMAXPROCS(0)
	var resultLock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < numProcs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			localCenterSum := make([]Vec, len(k.Centers))
			for i := range localCenterSum {
				localCenterSum[i] = make(Vec, dim)
			}
			localCenterCount := make([]int, len(k.Centers))
			localTotalError := 0.0
			for i := idx; i < len(k.Data); i += numProcs {
				co := k.Data[i]
				closestDist := 0.0
				closestIdx := 0
				for i, center := range k.Centers {
					d := float64(co.DistSquared(center))
					if d < closestDist || i == 0 {
						closestDist = d
						closestIdx = i
					}
				}
				localCenterSum[closestIdx] = localCenterSum[closestIdx].Add(co)
				localCenterCount[closestIdx]++
				localTotalError += closestDist
			}
			resultLock.Lock()
			defer resultLock.Unlock()
			for i, c := range localCenterCount {
				centerCount[i] += c
			}
			for i, s := range localCenterSum {
				centerSum[i] = centerSum[i].Add(s)
			}
			totalError += localTotalError
		}(i)
	}
	wg.Wait()

	for i, newCenter := range centerSum {
		count := centerCount[i]
		if count > 0 {
			k.Centers[i] = newCenter.Scale(1 / float64(count))
		}
	}

	return totalError / float64(len(k.Data))
}

// Assign gets the closest center index for every vector.
func (k *KMeans) Assign(vecs []Vec) []int {
	result := make([]int, len(vecs))
	essentials.ConcurrentMap(0, len(result), func(i int) {
		co := vecs[i]
		closestDist := 0.0
		closestIdx := 0
		for i, center := range k.Centers {
			d := co.DistSquared(center)
			if d < closestDist || i == 0 {
				closestDist = d
				closestIdx = i
			}
		}
		result[i] = closestIdx
	})
	return result
}

func kmeansPlusPlusInit(allColors []Vec, numCenters int) []Vec {
	centers := make([]Vec, numCenters)
	centers[0] = allColors[rand.Intn(len(allColors))]
	dists := newCenterDistances(allColors, centers[0])
	for i := 1; i < numCenters; i++ {
		sampleIdx := dists.Sample()
		centers[i] = allColors[sampleIdx]
		dists.Update(centers[i])
	}
	return centers
}

type centerDistances struct {
	Data        []Vec
	Distances   []float64
	DistanceSum float64
}

func newCenterDistances(data []Vec, center Vec) *centerDistances {
	dists := make([]float64, len(data))
	sum := 0.0
	for i, c := range data {
		dists[i] = float64(c.DistSquared(center))
		sum += dists[i]
	}
	return &centerDistances{
		Data:        data,
		Distances:   dists,
		DistanceSum: sum,
	}
}

func (c *centerDistances) Update(newCenter Vec) {
	c.DistanceSum = 0
	for i, co := range c.Data {
		d := float64(co.DistSquared(newCenter))
		if d < c.Distances[i] {
			c.Distances[i] = d
		}
		c.DistanceSum += c.Distances[i]
	}
}

func (c *centerDistances) Sample() int {
	sample := rand.Float64() * c.DistanceSum
	idx := len(c.Data) - 1
	for i, dist := range c.Distances {
		sample -= dist
		if sample < 0 {
			idx = i
			break
		}
	}
	return idx
}
