package utils

import (
	"crypto/rand"
	"encoding/binary"
	mathRand "math/rand"
	"time"
)

// RandInt generates a cryptographically secure random integer in range [0, max).
// Falls back to non-crypto random if crypto/rand fails (extremely rare).
func RandInt(maxValue int) int {
	if maxValue <= 0 {
		return 0
	}

	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		rng := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
		return rng.Intn(maxValue)
	}

	n := binary.BigEndian.Uint64(b[:])
	return int(n % uint64(maxValue))
}
