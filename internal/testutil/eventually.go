package testutil

import (
	"testing"
	"time"
)

// Eventually asserts that condition() becomes true within timeout.
// Polls every poll until it passes or the timeout expires.
//
// Useful for concurrent code where you need to wait for a goroutine
// to make observable progress without a tight sleep loop.
//
//	testutil.Eventually(t, time.Second, 10*time.Millisecond, func() bool {
//	    return atomic.LoadInt64(&counter) >= 5
//	}, "counter never reached 5")
func Eventually(tb testing.TB, timeout, poll time.Duration, condition func() bool, msg string) {
	tb.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return
		}
		if time.Now().After(deadline) {
			tb.Errorf("Eventually: %s (timeout=%v)", msg, timeout)
			return
		}
		time.Sleep(poll)
	}
}
