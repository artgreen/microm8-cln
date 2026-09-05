package files

import "testing"

// TestApple2IsHighCapacity documents the size/format threshold that the
// command-line -drive1 routing relies on to decide between the 140k Disk II
// controller and the SmartPort device (800k/400k 3.5" and hard-disk images).
func TestApple2IsHighCapacity(t *testing.T) {
	cases := []struct {
		name string
		ext  string
		size int
		want bool
	}{
		{"140k ProDOS .po -> Disk II", "po", 143360, false},
		{"140k DOS .dsk -> Disk II", "dsk", 143360, false},
		{"800k ProDOS .po -> SmartPort", "po", 819200, true},
		{"400k ProDOS .po -> SmartPort", "po", 409600, true},
		{".2mg always high capacity", "2mg", 143360, true},
		{".hdv always high capacity", "hdv", 143360, true},
		{"nib never high capacity", "nib", 819200, false},
		{"woz never high capacity", "woz", 819200, false},
	}
	for _, c := range cases {
		if got := Apple2IsHighCapacity(c.ext, c.size); got != c.want {
			t.Errorf("%s: Apple2IsHighCapacity(%q, %d) = %v, want %v",
				c.name, c.ext, c.size, got, c.want)
		}
	}
}
