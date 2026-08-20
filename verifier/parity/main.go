package main

import (
	"fmt"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func main() {
	vectors := [][2]float64{
		{1, 0}, {1, 100}, {2, 0}, {3, 50}, {5, 0}, {5, 50}, {5, 100},
		{8, 100}, {10, 0}, {10, 80}, {10, 100}, {15, 100}, {20, 90},
	}
	for _, v := range vectors {
		cc := int(v[0])
		fmt.Printf("%d %.1f %.12f\n", cc, v[1], crap.ComputeCRAP(cc, v[1]))
	}
}
