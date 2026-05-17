package disk

import (
	"testing"
)

// TestVTOC_HeaderAccessors exercises every field-reader on VTOC by
// constructing a hand-built sector with known values at the DOS 3.3
// header offsets. If anyone shifts an offset or changes a byte width,
// the corresponding accessor regresses here.
func TestVTOC_HeaderAccessors(t *testing.T) {
	t.Parallel()

	// Build a 256-byte sector matching the DOS 3.3 VTOC layout.
	// Layout reference (sector $11/$00):
	//   $01 / $02 — first catalog T/S
	//   $03      — DOS version
	//   $06      — volume number
	//   $27      — max T/S pairs per sector (122 = 0x7A)
	//   $31      — track allocation direction
	//   $34      — total tracks
	//   $35      — total sectors per track
	//   $36/$37  — bytes per sector (little-endian)
	var data [256]byte
	data[0x01] = 0x11 // catalog track 17
	data[0x02] = 0x0F // catalog sector 15
	data[0x03] = 0x03 // DOS 3.3
	data[0x06] = 0xFE // volume 254
	data[0x27] = 0x7A // 122 T/S pairs per sector
	data[0x31] = 0x01 // forward
	data[0x34] = 0x23 // 35 tracks
	data[0x35] = 0x10 // 16 sectors per track
	data[0x36] = 0x00 // bytes per sector low
	data[0x37] = 0x01 // bytes per sector high → 256

	v := &VTOC{}
	v.SetData(data[:], 17, 0)

	if cT, cS := v.GetCatalogStart(); cT != 17 || cS != 15 {
		t.Errorf("GetCatalogStart() = (%d,%d), want (17,15)", cT, cS)
	}
	if got := v.GetDOSVersion(); got != 0x03 {
		t.Errorf("GetDOSVersion = %#x, want 0x03", got)
	}
	if got := v.GetVolumeID(); got != 0xFE {
		t.Errorf("GetVolumeID = %#x, want 0xFE", got)
	}
	if got := v.GetMaxTSPairsPerSector(); got != 122 {
		t.Errorf("GetMaxTSPairsPerSector = %d, want 122", got)
	}
	if got := v.GetTrackOrder(); got != 1 {
		t.Errorf("GetTrackOrder = %d, want 1", got)
	}
	if got := v.GetTracks(); got != 35 {
		t.Errorf("GetTracks = %d, want 35", got)
	}
	if got := v.GetSectors(); got != 16 {
		t.Errorf("GetSectors = %d, want 16", got)
	}
	if got := v.BytesPerSector(); got != 256 {
		t.Errorf("BytesPerSector = %d, want 256", got)
	}
}

// TestVTOC_TrackSectorBitmap_RoundTrip exercises SetTSFree/IsTSFree at
// every track-sector boundary. The bitmap layout is tricky (per-track
// 4 bytes, sectors 0–7 in offset+1, sectors 8–15 in offset+0) — a
// regression in either index would flip the wrong bit and corrupt
// disk usage tracking.
func TestVTOC_TrackSectorBitmap_RoundTrip(t *testing.T) {
	t.Parallel()
	v := &VTOC{}
	v.SetData(make([]byte, 256), 17, 0)

	cases := []struct {
		track  int
		sector int
	}{
		{0, 0}, {0, 7}, {0, 8}, {0, 15}, // first track, every region
		{17, 0}, {17, 15}, // mid-disk
		{34, 0}, {34, 15}, // last track
	}

	// Initially all bits zero → not free.
	for _, tc := range cases {
		if v.IsTSFree(tc.track, tc.sector) {
			t.Errorf("T%d S%d reports free with zeroed bitmap", tc.track, tc.sector)
		}
	}

	// Set every test sector free, confirm the bit toggled — and
	// confirm we didn't accidentally toggle a neighbour.
	for _, tc := range cases {
		v.SetTSFree(tc.track, tc.sector, true)
		if !v.IsTSFree(tc.track, tc.sector) {
			t.Errorf("after SetTSFree(T%d S%d, true): IsTSFree = false",
				tc.track, tc.sector)
		}
	}

	// Clear them, confirm IsTSFree flips back.
	for _, tc := range cases {
		v.SetTSFree(tc.track, tc.sector, false)
		if v.IsTSFree(tc.track, tc.sector) {
			t.Errorf("after SetTSFree(T%d S%d, false): IsTSFree = true",
				tc.track, tc.sector)
		}
	}
}

// TestVTOC_TrackSectorBitmap_OnlyTouchesTargetBit asserts setting one
// T/S free doesn't accidentally toggle any other bit in the bitmap.
// Cheap protection against off-by-one in the bit-arithmetic.
func TestVTOC_TrackSectorBitmap_OnlyTouchesTargetBit(t *testing.T) {
	t.Parallel()
	v := &VTOC{}
	v.SetData(make([]byte, 256), 17, 0)

	v.SetTSFree(10, 5, true)

	for trk := 0; trk < 35; trk++ {
		for sec := 0; sec < 16; sec++ {
			free := v.IsTSFree(trk, sec)
			want := trk == 10 && sec == 5
			if free != want {
				t.Errorf("after SetTSFree(10, 5, true): IsTSFree(%d, %d) = %v, want %v",
					trk, sec, free, want)
			}
		}
	}
}
