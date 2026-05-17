// Package lifecycle holds small primitives for shaping long-lived
// goroutines so they exit cleanly on context cancellation.
//
// The pattern this replaces is the ubiquitous:
//
//	for {
//	    time.Sleep(d)
//	    doWork()
//	}
//
// which leaks one goroutine per call forever — there's no way to ask it to
// stop. With this package the equivalent is:
//
//	for lifecycle.SleepOrDone(ctx, d) {
//	    doWork()
//	}
//
// and the goroutine exits as soon as `ctx` is canceled.
package lifecycle

import (
	"context"
	"time"
)

// SleepOrDone blocks for up to `d`, returning true on a normal wake-up and
// false if the context is canceled (or already canceled on entry). Use it
// as the condition of a for-loop in a service goroutine:
//
//	for lifecycle.SleepOrDone(ctx, time.Second) {
//	    work()
//	}
//
// SleepOrDone is cheap when ctx is never canceled: a single timer is
// created per call and stopped before return so we don't leak entries in
// the runtime's timer heap.
func SleepOrDone(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		// Even with d==0 we still honour cancellation, so callers can use
		// SleepOrDone(ctx, 0) as a cheap "still running?" check.
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
