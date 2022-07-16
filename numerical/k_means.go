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
type KMeans[T Vector[T]] struct {
	// Centers is the K-means cluster centers.
	Centers []T

	// Data stores all of the data points.
	Data []T
}

// NewKMeans creates a KMeans object with K-means++
// initialization.
func NewKMeans[T Vector[T]](data []T, numCenters int) *KMeans[T] {
	if len(data) <= numCenters {
		return &KMeans[T]{
			Centers: data,
			Data:    data,
		}
	}
	return &KMeans[T]{
		Centers: kmeansPlusPlusInit(data, numCenters),
		Data:    data,
	}
}

// Iterate performs a step of k-means and returns the
// current MSE loss.
// If the MSE loss does not decrease, then the process has
// converged.
func (k *KMeans[T]) Iterate() float64 {
	zeroValue := k.Centers[0].Zeros()
	centerSum := make([]T, len(k.Centers))
	for i := range centerSum {
		centerSum[i] = zeroValue
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
			localCenterSum := make([]T, len(k.Centers))
			for i := range localCenterSum {
				localCenterSum[i] = zeroValue
			}
			localCenterCount := make([]int, len(k.Centers))
			localTotalError := 0.0
			for i := idx; i < len(k.Data); i += numProcs {
				co := k.Data[i]
				closestDist, closestIdx := k.assign(co)
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
func (k *KMeans[T]) Assign(vecs []T) []int {
	result := make([]int, len(vecs))
	essentials.ConcurrentMap(0, len(result), func(i int) {
		_, result[i] = k.assign(vecs[i])
	})
	return result
}

func (k *KMeans[T]) assign(v T) (dist float64, idx int) {
	for i, center := range k.Centers {
		d := float64(v.DistSquared(center))
		if d < dist || i == 0 {
			dist = d
			idx = i
		}
	}
	return
}

func kmeansPlusPlusInit[T Vector[T]](allColors []T, numCenters int) []T {
	centers := make([]T, numCenters)
	centers[0] = allColors[rand.Intn(len(allColors))]
	dists := newCenterDistances(allColors, centers[0])
	for i := 1; i < numCenters; i++ {
		sampleIdx := dists.Sample()
		centers[i] = allColors[sampleIdx]
		dists.Update(centers[i])
	}
	return centers
}

type centerDistances[T Vector[T]] struct {
	Data        []T
	Distances   []float64
	DistanceSum float64
}

func newCenterDistances[T Vector[T]](data []T, center T) *centerDistances[T] {
	dists := make([]float64, len(data))
	sum := 0.0
	for i, c := range data {
		dists[i] = float64(c.DistSquared(center))
		sum += dists[i]
	}
	return &centerDistances[T]{
		Data:        data,
		Distances:   dists,
		DistanceSum: sum,
	}
}

func (c *centerDistances[T]) Update(newCenter T) {
	c.DistanceSum = 0
	for i, co := range c.Data {
		d := float64(co.DistSquared(newCenter))
		if d < c.Distances[i] {
			c.Distances[i] = d
		}
		c.DistanceSum += c.Distances[i]
	}
}

func (c *centerDistances[T]) Sample() int {
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
