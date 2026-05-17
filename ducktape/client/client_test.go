package client

import (
	"net"
	"testing"
	"time"

	"paleotronic.com/ducktape"
	"paleotronic.com/internal/testutil"
)

// TestClientSender_QuitExitsLoop pins the Phase 5 SA4011 fix on the
// client-side ClientSender. The original `case <-client.Quit:` body
// ended with a bare `break` which only escaped the select; the
// surrounding `for client.OK { }` then re-entered the select. The fix
// sets `client.OK = false` so the for-loop's condition flips to false
// on the next iteration and the goroutine returns.
//
// We hand the goroutine a real net.Pipe end (so client.Conn.Close()
// works), signal Quit, and assert the goroutine returns within a
// short window.
func TestClientSender_QuitExitsLoop(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	clientConn, peer := net.Pipe()
	defer peer.Close()

	c := &DuckTapeClient{
		Name:     "test",
		Conn:     clientConn,
		Outgoing: make(chan *ducktape.DuckTapeBundle, 1),
		Quit:     make(chan bool, 1),
		OK:       true,
	}
	// ClientSender only reads .Outgoing / .Quit / .OK / .Conn from
	// the receiver; we don't need a full constructed client.

	done := make(chan struct{})
	go func() {
		c.ClientSender()
		close(done)
	}()

	// Let it run one tick.
	time.Sleep(10 * time.Millisecond)
	c.Quit <- true

	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Fatal("ClientSender did not exit within 1s of Quit; goroutine leak " +
			"regression (Phase 5 SA4011)")
	}
}
