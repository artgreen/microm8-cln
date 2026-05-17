package utils

import (
	"testing"
)

// TestRandom_ReturnsInUnitInterval is the basic happy-path: Random
// must produce values in [0, 1). Doesn't compare sequences so can
// run in parallel with the concurrency test.
func TestRandom_ReturnsInUnitInterval(t *testing.T) {
	t.Parallel()
	for i := 0; i < 1000; i++ {
		x := Random()
		if x < 0 || x >= 1 {
			t.Fatalf("Random() = %v, want in [0, 1)", x)
		}
	}
}

// TestPSeed_DeterministicSequence pins the Phase 2 rand modernization
// contract: PSeed(v) starts a deterministic sequence. Used by
// Applesoft's RND(<negative>) form. Before the modernization this was
// done via rand.Seed() (Go-1.20-deprecated, package-global); the fix
// rebinds the package-local *rand.Rand.
//
// Sequential — touches the package-global rng so cannot share the
// runtime with other tests calling Random().
func TestPSeed_DeterministicSequence(t *testing.T) {
	const seed = int64(42)
	PSeed(seed)
	first := []float64{Random(), Random(), Random(), Random(), Random()}

	PSeed(seed)
	second := []float64{Random(), Random(), Random(), Random(), Random()}

	for i := range first {
		if first[i] != second[i] {
			t.Errorf("PSeed(%d) sequence not deterministic: [%d] first=%v second=%v",
				seed, i, first[i], second[i])
		}
	}
}

// TestSeedRandom_ProducesNewSequence: SeedRandom() reseeds from the
// wall clock, so two calls separated by a moment should produce
// distinct subsequent values with overwhelming probability.
func TestSeedRandom_ProducesNewSequence(t *testing.T) {
	// Sequential — can't be parallel because we touch the global rng.
	PSeed(1)
	a := Random()
	SeedRandom() // reseeds from time.Now
	b := Random()
	if a == b {
		t.Errorf("Random() collided across SeedRandom(): %v == %v (effectively impossible)",
			a, b)
	}
}

// TestRandom_ConcurrentRaceClean is the Phase 8 follow-up: the
// package-local *rand.Rand is not thread-safe, so the Phase 2
// modernization needed a mutex. Run this under -race; if anyone
// removes the rngMu guard, this test fires.
func TestRandom_ConcurrentRaceClean(t *testing.T) {
	t.Parallel()
	const goroutines = 8
	const iters = 500

	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iters; j++ {
				_ = Random()
				if j%50 == 0 {
					_ = RandStringRunes(4)
				}
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
