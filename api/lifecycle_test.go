package s8webclient

import (
	"context"
	"sync"
	"testing"
	"time"

	"paleotronic.com/internal/testutil"
)

// newTestClient returns a Client configured for snappy tests: empty (no
// underlying DuckTape connection) and with a 5ms poll interval, so the
// monitor loop ticks several times per test rather than once every 10s.
func newTestClient() *Client {
	return &Client{MonitorPollInterval: 5 * time.Millisecond}
}

// TestMonitorNetwork_ExitsOnContextCancel is the load-bearing Phase-7
// regression test: the supervisor used to be `for { time.Sleep(10s); ... }`
// with no exit path, so a Client.Done() would leave the goroutine wedged
// forever. Now it exits when its context is canceled.
//
// We test against a Client with no underlying DuckTape connection (the
// reconnect branch becomes a no-op when c.c is nil). Cancel and confirm
// the goroutine returns within a generous timeout.
func TestMonitorNetwork_ExitsOnContextCancel(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	ctx, cancel := context.WithCancel(context.Background())
	c := newTestClient()

	done := make(chan struct{})
	go func() {
		MonitorNetwork(ctx, c)
		close(done)
	}()

	// Let it tick once or twice before cancelling.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// good — MonitorNetwork honoured cancellation
	case <-time.After(time.Second):
		t.Fatal("MonitorNetwork did not exit within 1s of cancellation")
	}
}

// TestClient_DoneCancelsMonitor confirms the wire-up between Client.Done
// and the monitor goroutine started by StartMonitor. Done() must cancel
// the stored context so the loop exits.
func TestClient_DoneCancelsMonitor(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	c := newTestClient()
	c.StartMonitor()

	// Snapshot ctx before Done(): Done() clears the field as it cancels.
	c.mu.Lock()
	ctx := c.monitorCtx
	c.mu.Unlock()
	if ctx == nil {
		t.Fatal("StartMonitor did not populate monitorCtx")
	}

	// Done() touches c.c when non-nil; ours is nil, which exercises the
	// safe-shutdown path.
	c.Done()

	select {
	case <-ctx.Done():
		// good — Done() cancelled the context
	case <-time.After(time.Second):
		t.Fatal("Client.Done did not cancel the monitor context")
	}
}

// TestClient_StartMonitorReplaceOldOne guards the "double-start" path:
// calling StartMonitor twice must cancel the first goroutine before
// launching the second, so we never end up with two supervisors.
func TestClient_StartMonitorReplaceOldOne(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	c := newTestClient()
	c.StartMonitor()

	c.mu.Lock()
	firstCtx := c.monitorCtx
	c.mu.Unlock()

	c.StartMonitor() // replace

	select {
	case <-firstCtx.Done():
		// good — first monitor was cancelled
	case <-time.After(time.Second):
		t.Fatal("second StartMonitor did not cancel the first context")
	}

	// Tear down the surviving monitor so the leak detector is happy.
	c.Done()
}

// TestMonitorNetwork_NilClientNoop is a defensive check: if a future
// caller passes a nil client (e.g. teardown ordering bug), the loop should
// keep running until cancellation rather than panicking on the nil deref.
// With a nil client we fall back to DefaultMonitorPollInterval (10s) which
// would make the test slow; instead we cancel quickly and rely on
// SleepOrDone honouring the cancel on the very first tick.
func TestMonitorNetwork_NilClientNoop(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MonitorNetwork(nil) panicked: %v", r)
			}
		}()
		MonitorNetwork(ctx, nil)
	}()

	// Cancel before the first 10s tick fires. SleepOrDone returns false
	// on the canceled context inside its select, so the loop exits cleanly.
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Fatal("MonitorNetwork(nil) did not exit on cancel")
	}
}
