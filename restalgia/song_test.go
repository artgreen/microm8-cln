//go:build ignore

// Stale — NewSong's signature changed from `NewSong(string) *Song` to
// `NewSong(string) (*Song, error)`. Quarantined until rewritten or deleted.
// TODO(modernize/phase-4): decide fate.

package restalgia

import (
	"testing"
)

func TestKeypressPackUnpack(t *testing.T) {

	song := NewSong("test.song")

	////fmt.Printntln(song)

	s := 0
	for song.Playing {
		if song.PullSampleMono() != 0 {
			//////fmt.Println(v)
			s++
		}
	}

	////fmt.Printntf("Samples non zero = %d\n", s)

}
