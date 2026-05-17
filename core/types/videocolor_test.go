package types

import (
	"math"
	"testing"
)

// TestVideoColor_LabConstantsAreFractional pins the Phase 5 SA4025 fix:
// the CIE Lab eps/k thresholds were written as `216 / 24389` and
// `24389 / 27` — integer division, evaluating to 0 and 904 respectively
// instead of the real CIE constants 0.008856… and 903.296…. The bug
// silently corrupted every Lab colour conversion in the palette code.
//
// Note: rgb2lab() returns a packed/scaled L (it multiplies by 2.55 and
// offsets by 0.5, so L ranges roughly [0.5, 255.5] rather than the
// standard [0, 100]). We test against that observed scale.
//
// Three regression catches:
//
//  1. Near-black (RGB 1,1,1) takes the LINEAR branch of the
//     XYZ→Lab transform because yr ≈ 3.3e-4 < eps ≈ 0.0089. Under
//     the bug, eps=0 means every input takes the cube-root branch,
//     producing L < 0 for near-black. The fix produces L > 0.
//  2. Pure black returns L=0 before scale → 0.5 after the +0.5.
//  3. Pure white returns L=100 before scale → ≈255.5 after.
func TestVideoColor_LabConstantsAreFractional(t *testing.T) {
	t.Parallel()

	// Pure black: pre-scale L=0, post-scale 0.5.
	L, _, _ := rgb2lab(0, 0, 0)
	if math.Abs(L-0.5) > 0.01 {
		t.Errorf("rgb2lab(0,0,0) L = %v, want ≈0.5 (post-scale)", L)
	}

	// Pure white: pre-scale L=100, post-scale ≈255.5.
	L, _, _ = rgb2lab(255, 255, 255)
	if math.Abs(L-255.5) > 1.0 {
		t.Errorf("rgb2lab(255,255,255) L = %v, want ≈255.5 (post-scale)", L)
	}

	// Near-black (R=G=B=1) — THE bug-pinning case. Under SA4025 with
	// eps=0 this would take the cube-root branch and produce a negative
	// L. Under the fix it takes the linear branch and produces a small
	// positive L. The assertion "L ≥ 0" is what distinguishes the two.
	L, _, _ = rgb2lab(1, 1, 1)
	if L < 0 {
		t.Errorf("rgb2lab(1,1,1) L = %v < 0; integer-divide bug in eps "+
			"would force the cube-root branch and produce negative L "+
			"(Phase 5 SA4025 regression)", L)
	}
}

// TestVideoColor_ToCIELAB_ConsistentWithRgb2lab is a thin wrapper-level
// sanity check: the public ToCIELAB() must agree with rgb2lab() for
// the same RGB. If someone refactors VideoColor to compute Lab inline
// and forgets a constant, this catches it.
func TestVideoColor_ToCIELAB_ConsistentWithRgb2lab(t *testing.T) {
	t.Parallel()
	vc := &VideoColor{Red: 100, Green: 150, Blue: 200, Alpha: 255}
	wantL, wantA, wantB := rgb2lab(100, 150, 200)
	gotL, gotA, gotB := vc.ToCIELAB()
	if gotL != wantL || gotA != wantA || gotB != wantB {
		t.Errorf("ToCIELAB() = (%v,%v,%v), rgb2lab = (%v,%v,%v)",
			gotL, gotA, gotB, wantL, wantA, wantB)
	}
}
