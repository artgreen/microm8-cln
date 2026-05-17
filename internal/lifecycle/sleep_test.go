package lifecycle

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSleepOrDone_NormalWake(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	got := SleepOrDone(ctx, 25*time.Millisecond)
	elapsed := time.Since(start)

	if !got {
		t.Errorf("SleepOrDone returned false without cancellation")
	}
	if elapsed < 20*time.Millisecond {
		// scheduler jitter cushion of 5ms — be generous on CI
		t.Errorf("SleepOrDone returned too early: %v (wanted ≥20ms)", elapsed)
	}
}

func TestSleepOrDone_CancelDuringSleep(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel halfway through the would-be sleep.
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	got := SleepOrDone(ctx, time.Second)
	elapsed := time.Since(start)

	if got {
		t.Errorf("SleepOrDone returned true after cancellation")
	}
	// Should observe cancel within a couple-ish of scheduler ticks.
	if elapsed > 100*time.Millisecond {
		t.Errorf("SleepOrDone took too long to observe cancel: %v", elapsed)
	}
}

func TestSleepOrDone_AlreadyCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Even a long duration must return immediately when ctx is already done.
	start := time.Now()
	got := SleepOrDone(ctx, time.Hour)
	elapsed := time.Since(start)

	if got {
		t.Errorf("SleepOrDone returned true on pre-canceled context")
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("SleepOrDone slow on pre-canceled context: %v", elapsed)
	}
}

func TestSleepOrDone_ZeroDuration(t *testing.T) {
	t.Parallel()
	t.Run("alive", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		if !SleepOrDone(ctx, 0) {
			t.Errorf("zero-duration on live ctx should return true")
		}
	})
	t.Run("canceled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if SleepOrDone(ctx, 0) {
			t.Errorf("zero-duration on canceled ctx should return false")
		}
	})
}

// TestSleepOrDone_LoopExits exercises the canonical use site: a service
// loop using SleepOrDone as the condition. The goroutine must exit when
// the context is canceled. This is the property that turns a leaked
// goroutine into a managed one.
func TestSleepOrDone_LoopExits(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticks := 0
		for SleepOrDone(ctx, 5*time.Millisecond) {
			ticks++
			if ticks > 1000 {
				// safety belt: shouldn't reach here under cancel
				return
			}
		}
	}()

	// Let it tick a few times to confirm the loop is alive, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// goroutine exited
	case <-time.After(time.Second):
		t.Fatal("loop did not exit within 1s of cancellation")
	}
}
