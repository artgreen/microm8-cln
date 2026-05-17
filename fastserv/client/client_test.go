package client

import (
	"net"
	"testing"
	"time"

	"paleotronic.com/internal/testutil"
)

// TestClientSender_QuitExitsLoop pins the Phase 5 SA4011 fix on the
// fastserv ClientSender. The original Quit case ended with a bare
// `break` that only escaped the select; the for-loop's
// `for client.OK` condition would not flip false until a subsequent
// write failed, so the goroutine spun. The fix adds
// `client.OK = false` inside the Quit case so the for-loop exits on
// the next iteration.
func TestClientSender_QuitExitsLoop(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	clientConn, peer := net.Pipe()
	defer peer.Close()

	c := &FSClient{
		Name:     "test",
		Conn:     clientConn,
		Outgoing: make(chan []byte, 1),
		Quit:     make(chan bool, 1),
		OK:       true,
	}

	done := make(chan struct{})
	go func() {
		c.ClientSender()
		close(done)
	}()

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
