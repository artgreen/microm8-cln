package disk

import (
	"testing"
)

// TestSectorMapperLinear is the trivial baseline: identity for every
// sector index 0–15.
func TestSectorMapperLinear(t *testing.T) {
	t.Parallel()
	for i := 0; i <= 15; i++ {
		if got := SectorMapperLinear(i); got != i {
			t.Errorf("SectorMapperLinear(%d) = %d, want %d", i, got, i)
		}
	}
}

// TestSectorMapperDOS33_Identity pins the current DOS 3.3 mapping.
// After the Phase 4c sweep this is the identity function — the
// original interleave table was unreachable behind an early return.
// If anyone re-introduces the interleave without checking what
// callers expect, this fails and points at the regression.
func TestSectorMapperDOS33_Identity(t *testing.T) {
	t.Parallel()
	for i := 0; i <= 15; i++ {
		if got := SectorMapperDOS33(i); got != i {
			t.Errorf("SectorMapperDOS33(%d) = %d, want %d (currently identity post-Phase-4c)",
				i, got, i)
		}
	}
}

// TestSectorMapperDOS33Alt_IsBijection asserts the DOS33Alt mapping
// is a permutation of 0..15. A mistake in the switch (duplicate
// target, missing entry) would surface here as a missing or repeated
// output.
func TestSectorMapperDOS33Alt_IsBijection(t *testing.T) {
	t.Parallel()
	seen := make(map[int]int)
	for i := 0; i <= 15; i++ {
		out := SectorMapperDOS33Alt(i)
		if out < 0 || out > 15 {
			t.Errorf("SectorMapperDOS33Alt(%d) = %d, want in [0, 15]", i, out)
			continue
		}
		if prev, dup := seen[out]; dup {
			t.Errorf("SectorMapperDOS33Alt: %d and %d both map to %d", prev, i, out)
		}
		seen[out] = i
	}
	if len(seen) != 16 {
		t.Errorf("SectorMapperDOS33Alt covers %d outputs, want 16", len(seen))
	}
}

// TestSectorMapperDOS33Alt_SpotChecks pins specific values from the
// switch. If anyone "simplifies" the function and breaks the
// translation, the spot-checks scream.
func TestSectorMapperDOS33Alt_SpotChecks(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want int }{
		{0, 0},   // first
		{15, 15}, // last
		{13, 1},  // mid
		{1, 7},   // mid
		{6, 12},  // mid
	}
	for _, tc := range cases {
		if got := SectorMapperDOS33Alt(tc.in); got != tc.want {
			t.Errorf("SectorMapperDOS33Alt(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestSectorMapperDOS33Alt_Invalid: out-of-range inputs return -1
// (the documented "invalid sector" sentinel).
func TestSectorMapperDOS33Alt_Invalid(t *testing.T) {
	t.Parallel()
	for _, bad := range []int{-1, 16, 100, 1 << 20} {
		if got := SectorMapperDOS33Alt(bad); got != -1 {
			t.Errorf("SectorMapperDOS33Alt(%d) = %d, want -1 (invalid sentinel)", bad, got)
		}
	}
}

// TestSectorMapperProDOS_IsBijection asserts the ProDOS mapping is a
// permutation of 0..15 — same property as DOS33Alt, different table.
func TestSectorMapperProDOS_IsBijection(t *testing.T) {
	t.Parallel()
	seen := make(map[int]int)
	for i := 0; i <= 15; i++ {
		out := SectorMapperProDOS(i)
		if out < 0 || out > 15 {
			t.Errorf("SectorMapperProDOS(%d) = %d, want in [0, 15]", i, out)
			continue
		}
		if prev, dup := seen[out]; dup {
			t.Errorf("SectorMapperProDOS: %d and %d both map to %d", prev, i, out)
		}
		seen[out] = i
	}
	if len(seen) != 16 {
		t.Errorf("SectorMapperProDOS covers %d outputs, want 16", len(seen))
	}
}

// TestSectorMapperProDOS_SpotChecks pins specific values from the
// switch table.
func TestSectorMapperProDOS_SpotChecks(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want int }{
		{0, 0},   // identity at both ends
		{15, 15}, // identity at both ends
		{1, 2},   // first non-identity
		{7, 14},  // ladder peak
		{8, 1},   // ladder restart
		{14, 13}, // tail
	}
	for _, tc := range cases {
		if got := SectorMapperProDOS(tc.in); got != tc.want {
			t.Errorf("SectorMapperProDOS(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestSectorMapperProDOS_Invalid: out-of-range returns -1.
func TestSectorMapperProDOS_Invalid(t *testing.T) {
	t.Parallel()
	for _, bad := range []int{-1, 16, 999} {
		if got := SectorMapperProDOS(bad); got != -1 {
			t.Errorf("SectorMapperProDOS(%d) = %d, want -1", bad, got)
		}
	}
}

// TestSectorMapperDiversiDOS_IsBijection_Or_Invalid: DiversiDOS uses
// hex constants 0x00-0x0f. Confirm it's a bijection on [0,15] or
// returns -1 outside.
func TestSectorMapperDiversiDOS_IsBijection(t *testing.T) {
	t.Parallel()
	seen := make(map[int]int)
	for i := 0; i <= 15; i++ {
		out := SectorMapperDiversiDOS(i)
		if out < 0 || out > 15 {
			// Some valid sources may legitimately return -1, but the
			// expected behaviour is a 0..15 permutation.
			t.Errorf("SectorMapperDiversiDOS(%d) = %d, want in [0, 15]", i, out)
			continue
		}
		if prev, dup := seen[out]; dup {
			t.Errorf("SectorMapperDiversiDOS: %d and %d both map to %d", prev, i, out)
		}
		seen[out] = i
	}
}
