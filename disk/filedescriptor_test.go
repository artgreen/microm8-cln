package disk

import (
	"testing"
)

// newFD builds a FileDescriptor backed by a freshly-allocated 35-byte
// buffer so the SetXxx setters have somewhere to write. Bytes are
// laid out per the AppleDOS catalog-entry format:
//
//	$00      T of first T/S list sector
//	$01      S of first T/S list sector
//	$02      Type & locked flag (high bit = locked)
//	$03-$20  30-byte filename (high-ASCII, space-padded)
//	$21-$22  total sector count (little-endian)
func newFD(t *testing.T) *FileDescriptor {
	t.Helper()
	fd := &FileDescriptor{Data: make([]byte, 35)}
	return fd
}

// TestFileDescriptor_NameSetGetRoundTrip pins the AppleDOS filename
// encoding. SetName writes high-ASCII (byte | 0x80) and right-pads
// with 0xA0 (space | 0x80). Name() decodes back to lowercase trimmed.
func TestFileDescriptor_NameSetGetRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		set  string
		want string
	}{
		{"HELLO", "hello"},
		{"GAME", "game"},
		{"A", "a"},
		// 30-char max — SetName truncates beyond that.
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ12345", "abcdefghijklmnopqrstuvwxyz1234"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.set, func(t *testing.T) {
			t.Parallel()
			fd := newFD(t)
			fd.SetName(tc.set)
			if got := fd.NameUnadorned(); got != tc.want {
				t.Errorf("SetName(%q) → NameUnadorned() = %q, want %q",
					tc.set, got, tc.want)
			}
		})
	}
}

// TestFileDescriptor_Name_AppendsExtensionByType pins the
// type-suffix contract: Name() returns "<base>.a" for Applesoft,
// ".i" for Integer BASIC, ".s" for binary, ".t" for text. The
// filename-extension routing in dialect dispatch depends on this.
func TestFileDescriptor_Name_AppendsExtensionByType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		typ  FileType
		want string
	}{
		{FileTypeAPP, "hello.a"},
		{FileTypeINT, "hello.i"},
		{FileTypeBIN, "hello.s"},
		{FileTypeTXT, "hello.t"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			fd := newFD(t)
			fd.SetName("HELLO")
			fd.SetType(tc.typ)
			if got := fd.Name(); got != tc.want {
				t.Errorf("Name() for type %v = %q, want %q", tc.typ, got, tc.want)
			}
		})
	}
}

// TestFileDescriptor_LockedBit pins the type/locked overlap: type
// occupies bits 0-6 of byte $02; bit 7 is the locked flag. Setting
// one mustn't disturb the other.
func TestFileDescriptor_LockedBit(t *testing.T) {
	t.Parallel()
	fd := newFD(t)

	fd.SetType(FileTypeBIN)
	if fd.IsLocked() {
		t.Error("freshly-set type already reports locked")
	}

	fd.SetLocked(true)
	if !fd.IsLocked() {
		t.Error("after SetLocked(true): IsLocked = false")
	}
	if fd.Type() != FileTypeBIN {
		t.Errorf("Type drifted after SetLocked(true): got %v, want FileTypeBIN", fd.Type())
	}

	fd.SetLocked(false)
	if fd.IsLocked() {
		t.Error("after SetLocked(false): IsLocked = true")
	}
	if fd.Type() != FileTypeBIN {
		t.Errorf("Type drifted after SetLocked(false): got %v", fd.Type())
	}
}

// TestFileDescriptor_TotalSectors_LittleEndian asserts the 16-bit
// sector count is stored low-byte/high-byte at $21/$22. A regression
// here would silently report wrong file sizes.
func TestFileDescriptor_TotalSectors_LittleEndian(t *testing.T) {
	t.Parallel()
	fd := newFD(t)
	for _, want := range []int{0, 1, 255, 256, 0x1234, 0xFFFF} {
		fd.SetTotalSectors(want)
		if got := fd.TotalSectors(); got != want {
			t.Errorf("SetTotalSectors(%d) → TotalSectors() = %d", want, got)
		}
	}
}

// TestFileDescriptor_TrackSectorListStart_RoundTrip pins the T/S
// list pointer at offsets $00/$01.
func TestFileDescriptor_TrackSectorListStart_RoundTrip(t *testing.T) {
	t.Parallel()
	fd := newFD(t)
	for _, want := range []struct{ t, s int }{
		{0, 0},
		{17, 15},
		{34, 0},
		{255, 255},
	} {
		fd.SetTrackSectorListStart(want.t, want.s)
		gT, gS := fd.GetTrackSectorListStart()
		if gT != want.t || gS != want.s {
			t.Errorf("SetTrackSectorListStart(%d,%d) → Get = (%d,%d)",
				want.t, want.s, gT, gS)
		}
	}
}

// TestFileDescriptor_IsUnused: an entry with data[0] == 0xFF (the
// AppleDOS "never used" sentinel) reports unused. A regression here
// would cause catalog walkers to skip live entries or include garbage.
func TestFileDescriptor_IsUnused(t *testing.T) {
	t.Parallel()
	fd := newFD(t)
	fd.SetName("REAL")
	fd.SetType(FileTypeAPP)
	fd.SetTotalSectors(5)
	fd.SetTrackSectorListStart(20, 3)

	if fd.IsUnused() {
		t.Error("well-formed FD reports unused")
	}

	fd.Data[0] = 0xff
	if !fd.IsUnused() {
		t.Error("FD with 0xff sentinel doesn't report unused")
	}
}
