package server

import (
	"container/list"
	"net"
	"testing"
	"time"

	"paleotronic.com/ducktape"
	"paleotronic.com/internal/testutil"
)

// newTestClient builds a *ducktape.Client suitable for exercising the
// server's ClientSender goroutine: a real (closeable) net.Pipe end on
// Conn, in-memory channels, an empty client list.
func newTestClient(t *testing.T) (*ducktape.Client, net.Conn, *list.List) {
	t.Helper()
	clientConn, peer := net.Pipe()
	clientList := list.New()
	c := &ducktape.Client{
		Name:       "test",
		Conn:       clientConn,
		Incoming:   make(chan *ducktape.DuckTapeBundle, 1),
		Outgoing:   make(chan *ducktape.DuckTapeBundle, 1),
		Quit:       make(chan bool, 1),
		ClientList: clientList,
		Connected:  true,
	}
	return c, peer, clientList
}

// TestClientSender_QuitTerminatesGoroutine is the Phase 5 SA4011 pin.
// The original `ClientSender` had a bare `break` in the Quit case of
// its `select { ... }`, which only escaped the SELECT — the
// surrounding `for { }` had no exit condition, so the goroutine
// leaked forever even though Quit had been received. The fix
// replaces `break` with `return`.
//
// We hand the goroutine a Client with a real net.Pipe end (so
// client.Close() can call Conn.Close() without panicking) and a
// fresh list.List for RemoveMe. We then signal Quit and assert the
// goroutine returns within a couple of seconds (client.Close sleeps
// 1s before closing the conn).
func TestClientSender_QuitTerminatesGoroutine(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: client.Close sleeps 1s before tearing down the conn")
	}
	defer testutil.NoGoroutineLeaks(t)()

	c, peer, _ := newTestClient(t)
	defer peer.Close()

	srv := &DuckTapeServer{} // ClientSender doesn't read any server state

	done := make(chan struct{})
	go func() {
		srv.ClientSender(c)
		close(done)
	}()

	// Let the goroutine spin once on the `default` (sleep 1ms) branch.
	time.Sleep(10 * time.Millisecond)

	// Signal Quit.
	c.Quit <- true

	select {
	case <-done:
		// good — the goroutine returned cleanly
	case <-time.After(3 * time.Second):
		t.Fatal("ClientSender did not exit within 3s of Quit signal; " +
			"goroutine leak regression (Phase 5 SA4011)")
	}
}
