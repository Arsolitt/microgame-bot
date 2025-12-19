package utils

import (
	"math/rand/v2"
)

// RandInt generates a random integer in range [0, max).
func RandInt(maxValue int) int {
	if maxValue <= 0 {
		return 0
	}
	return rand.IntN(maxValue)
}
