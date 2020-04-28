package render3d

import (
	"math/rand"
	"runtime"
	"sync"
)

type goInfo struct {
	Gen   *rand.Rand
	Extra interface{}
}

// mapCoordinates calls f with every coordinate in an
// image, along with a per-goroutine random number
// generator and the pixel index.
func mapCoordinates(width, height int, f func(g *goInfo, x, y, idx int)) {
	coords := make(chan [3]int, width*height)
	var idx int
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			coords <- [3]int{x, y, idx}
			idx++
		}
	}
	close(coords)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g := &goInfo{
				Gen: rand.New(rand.NewSource(rand.Int63())),
			}
			for c := range coords {
				f(g, c[0], c[1], c[2])
			}
		}()
	}

	wg.Wait()
}
