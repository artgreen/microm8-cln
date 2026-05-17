package testutil

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

// NoGoroutineLeaks asserts that no extra goroutines are running at the end
// of the test compared to the start. Call near the top of any test that
// spawns goroutines — typically context-driven workers — to catch leaks
// when cancellation paths fail to wind down.
//
//	func TestWorker(t *testing.T) {
//	    defer testutil.NoGoroutineLeaks(t)()
//	    ...
//	}
//
// The returned function should be deferred (note the trailing `()`).
// It allows up to settleTimeout for spawned goroutines to exit before
// declaring a leak; the goroutine scheduler runs asynchronously so we
// can't assume immediate cleanup.
func NoGoroutineLeaks(tb testing.TB) func() {
	tb.Helper()
	const settleTimeout = 250 * time.Millisecond
	start := goroutineSnapshot()

	return func() {
		tb.Helper()
		// Allow newly-cancelled goroutines a moment to exit.
		deadline := time.Now().Add(settleTimeout)
		for {
			end := goroutineSnapshot()
			leaked := diffGoroutines(start, end)
			if len(leaked) == 0 {
				return
			}
			if time.Now().After(deadline) {
				tb.Errorf("goroutine leak: %d extra goroutine(s) at end of test:\n%s",
					len(leaked), strings.Join(leaked, "\n---\n"))
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func goroutineSnapshot() map[string]string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	stacks := strings.Split(string(buf[:n]), "\n\n")
	out := make(map[string]string, len(stacks))
	for _, s := range stacks {
		// First line: "goroutine N [state]:" — use N as key
		idx := strings.Index(s, " ")
		if idx <= 0 {
			continue
		}
		end := strings.Index(s[idx+1:], " ")
		if end < 0 {
			continue
		}
		id := s[idx+1 : idx+1+end]
		out[id] = s
	}
	return out
}

func diffGoroutines(before, after map[string]string) []string {
	var extra []string
	for id, stack := range after {
		if _, ok := before[id]; ok {
			continue
		}
		// Filter goroutines that are clearly Go runtime / test infrastructure.
		if strings.Contains(stack, "runtime.gopark") &&
			(strings.Contains(stack, "testing.(*T).Run") ||
				strings.Contains(stack, "runtime.goexit")) {
			// Heuristic: a single-frame parked goroutine in the test runner
			// itself isn't a user leak. Real leaks have meaningful stack frames.
			if strings.Count(stack, "\n") < 5 {
				continue
			}
		}
		extra = append(extra, stack)
	}
	return extra
}
