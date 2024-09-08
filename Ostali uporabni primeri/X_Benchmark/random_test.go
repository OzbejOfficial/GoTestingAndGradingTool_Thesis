package main

import (
	"math/rand"
	"testing"
)

func BenchmarkRandInt(b *testing.B) {
	for range b.N {
		rand.Int()
	}
}

// Primerjalni test
