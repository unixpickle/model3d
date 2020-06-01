package main

import "fmt"

func main() {
	segs := 0
	for i, d := range AllDigits() {
		segs += len(d)
		fmt.Println(i+1, len(d))
	}
	fmt.Println(segs)
}
