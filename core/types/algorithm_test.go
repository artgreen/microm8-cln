package types

import (
	"sync"
	"testing"
)

// TestAlgorithm_PutGetRoundTrip is the happy-path baseline. The other
// tests here drive concurrent access; this one guarantees the basic
// contract still works.
func TestAlgorithm_PutGetRoundTrip(t *testing.T) {
	t.Parallel()
	a := NewAlgorithm()
	line := Line{NewStatement()}
	a.Put(10, line)

	got, ok := a.Get(10)
	if !ok {
		t.Fatalf("Get(10) returned !ok after Put(10, ...)")
	}
	if len(got) != 1 {
		t.Fatalf("Get(10) len = %d, want 1", len(got))
	}

	if _, ok := a.Get(20); ok {
		t.Error("Get(20) returned ok for never-put key")
	}

	a.Remove(10)
	if _, ok := a.Get(10); ok {
		t.Error("Get(10) returned ok after Remove(10)")
	}
}

// TestAlgorithm_ConcurrentAccessUnderRace is the Phase 4b regression
// pin. Before that phase, Algorithm's methods (Size, Get, ContainsKey,
// PrevAfter, NextAfter, String, Checksum) had VALUE receivers — so
// `this.m.Lock()` was operating on a *copy* of the mutex, providing
// zero synchronisation. Concurrent callers raced on the C map.
//
// After the fix all methods use pointer receivers; the mutex actually
// protects the map. Run this under `-race`: if anyone regresses any of
// the methods to a value receiver, this test trips.
func TestAlgorithm_ConcurrentAccessUnderRace(t *testing.T) {
	t.Parallel()
	a := NewAlgorithm()

	// Seed with a few entries so Get/ContainsKey have something to read.
	for i := 1; i <= 5; i++ {
		a.Put(i*10, Line{NewStatement()})
	}

	const (
		writers = 4
		readers = 4
		iters   = 200
	)
	var wg sync.WaitGroup

	for w := 0; w < writers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				key := w*1000 + i
				a.Put(key, Line{NewStatement()})
				a.Remove(key)
			}
		}()
	}

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				_ = a.ContainsKey(10)
				_, _ = a.Get(20)
				_ = a.Size()
			}
		}()
	}

	wg.Wait()
}

// TestAlgorithm_GetSortedKeysReturnsDeterministicOrder pins the
// observable contract of GetSortedKeys: keys come back in ascending
// order. The lock-by-value bug would have surfaced here as map-order
// flakiness; the fix makes the result deterministic.
func TestAlgorithm_GetSortedKeysReturnsDeterministicOrder(t *testing.T) {
	t.Parallel()
	a := NewAlgorithm()
	for _, k := range []int{50, 10, 30, 20, 40} {
		a.Put(k, Line{NewStatement()})
	}
	keys := a.GetSortedKeys()
	want := []int{10, 20, 30, 40, 50}
	if len(keys) != len(want) {
		t.Fatalf("GetSortedKeys len = %d, want %d", len(keys), len(want))
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("GetSortedKeys[%d] = %d, want %d", i, k, want[i])
		}
	}
}
