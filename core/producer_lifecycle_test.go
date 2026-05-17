package core

import (
	"context"
	"testing"
	"time"

	"paleotronic.com/internal/testutil"
)

// Phase-7 regression tests: MusicService and RebootService used to be
// `for { time.Sleep(...); ... }` loops with no exit, so they leaked one
// goroutine per Producer forever. Now both honour a context.
//
// We test against a bare *Producer literal — we don't go through
// NewProducer because that constructs a real VM stack which needs the
// full apple2/glumby/restalgia subsystem available at link-time. The
// loop's exit behaviour is what we care about; the per-tick body is a
// no-op when there are no VMs.

// newTestProducer returns a Producer with snappy tick intervals so the
// service loops run several iterations per test rather than once every
// few hundred milliseconds.
func newTestProducer() *Producer {
	return &Producer{
		MusicServiceTick:  2 * time.Millisecond,
		RebootServiceTick: 2 * time.Millisecond,
	}
}

func TestMusicService_ExitsOnContextCancel(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	p := newTestProducer()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		p.MusicService(ctx)
		close(done)
	}()

	// Let it tick once or twice before cancelling.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Fatal("MusicService did not exit within 1s of cancellation")
	}
}

func TestRebootService_ExitsOnContextCancel(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	p := newTestProducer()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		p.RebootService(ctx)
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Fatal("RebootService did not exit within 1s of cancellation")
	}
}

// TestProducer_ShutdownCancelsServices is the high-level lifecycle test:
// StartServices launches both helpers, Shutdown stops them, and no
// goroutines are leaked.
func TestProducer_ShutdownCancelsServices(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	p := newTestProducer()
	p.StartServices()
	// Snapshot the ctx before Shutdown clears the field.
	ctx := p.servicesCtx
	if ctx == nil {
		t.Fatal("StartServices did not populate servicesCtx")
	}

	p.Shutdown()

	select {
	case <-ctx.Done():
		// good
	case <-time.After(time.Second):
		t.Fatal("Shutdown did not cancel the services context")
	}
}

// TestProducer_StartServicesReplacesPrior confirms a double-StartServices
// cancels the first generation, so we never accumulate supervisor pairs.
func TestProducer_StartServicesReplacesPrior(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	p := newTestProducer()
	p.StartServices()
	firstCtx := p.servicesCtx

	p.StartServices() // should cancel firstCtx and create a new one

	select {
	case <-firstCtx.Done():
		// good
	case <-time.After(time.Second):
		t.Fatal("second StartServices did not cancel the first context")
	}

	// Tear down the surviving generation.
	p.Shutdown()
}

// TestProducer_ShutdownIdempotent: calling Shutdown twice (or before
// StartServices) should be safe.
func TestProducer_ShutdownIdempotent(t *testing.T) {
	defer testutil.NoGoroutineLeaks(t)()

	p := newTestProducer()
	// Shutdown before StartServices — no-op, must not panic.
	p.Shutdown()

	p.StartServices()
	p.Shutdown()
	p.Shutdown() // second call — no-op, must not panic
}
