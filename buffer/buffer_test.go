package buffer

import (
	"testing"
)

// --- Behavior tests ----------------------------------------------------------

func TestRingBuffer_EmptyOnConstruction(t *testing.T) {
	t.Parallel()
	rb := NewRingBuffer(4, false)
	if !rb.Empty() {
		t.Fatal("newly constructed RingBuffer should be Empty()")
	}
	if v, ok := rb.Pop(); ok || v != nil {
		t.Fatalf("Pop on empty buffer: got (%v, %v), want (nil, false)", v, ok)
	}
}

func TestRingBuffer_FIFOOrder(t *testing.T) {
	t.Parallel()
	rb := NewRingBuffer(8, false)
	for i := 0; i < 5; i++ {
		if !rb.Push(i) {
			t.Fatalf("Push(%d) returned false unexpectedly", i)
		}
	}
	if rb.Empty() {
		t.Fatal("Empty() = true after 5 pushes, want false")
	}
	for i := 0; i < 5; i++ {
		v, ok := rb.Pop()
		if !ok {
			t.Fatalf("Pop() #%d returned ok=false unexpectedly", i)
		}
		if v.(int) != i {
			t.Errorf("Pop() #%d = %v, want %d (FIFO violation)", i, v, i)
		}
	}
	if !rb.Empty() {
		t.Errorf("Empty() = false after draining, want true")
	}
}

func TestRingBuffer_FullReturnsFalseWhenGrowDisabled(t *testing.T) {
	t.Parallel()
	const cap = 4
	rb := NewRingBuffer(cap, false)
	for i := 0; i < cap; i++ {
		if !rb.Push(i) {
			t.Fatalf("Push(%d) failed within capacity", i)
		}
	}
	// One more must be rejected.
	if rb.Push("overflow") {
		t.Error("Push beyond capacity returned true, want false")
	}
}

func TestRingBuffer_WrapAround(t *testing.T) {
	t.Parallel()
	const cap = 4
	rb := NewRingBuffer(cap, false)

	// Fill, drain half, refill — this exercises the wrap path.
	for i := 0; i < cap; i++ {
		rb.Push(i)
	}
	for i := 0; i < 2; i++ {
		v, _ := rb.Pop()
		if v.(int) != i {
			t.Fatalf("Pre-wrap Pop #%d = %v, want %d", i, v, i)
		}
	}
	// Two slots available; push two more.
	rb.Push(100)
	rb.Push(101)

	// Expected drain order: 2, 3, 100, 101
	want := []int{2, 3, 100, 101}
	for i, w := range want {
		v, ok := rb.Pop()
		if !ok {
			t.Fatalf("Post-wrap Pop #%d returned ok=false", i)
		}
		if v.(int) != w {
			t.Errorf("Post-wrap Pop #%d = %v, want %d", i, v, w)
		}
	}
}

func TestRingBuffer_GrowsWhenEnabled(t *testing.T) {
	t.Parallel()
	const initial = 4
	rb := NewRingBuffer(initial, true)
	const total = initial * 3

	for i := 0; i < total; i++ {
		if !rb.Push(i) {
			t.Fatalf("Push(%d) returned false with grow=true at iteration %d", i, i)
		}
	}
	for i := 0; i < total; i++ {
		v, ok := rb.Pop()
		if !ok {
			t.Fatalf("Pop #%d returned ok=false", i)
		}
		if v.(int) != i {
			t.Errorf("Post-grow Pop #%d = %v, want %d", i, v, i)
		}
	}
}

func TestRingBuffer_AllowsArbitraryTypes(t *testing.T) {
	t.Parallel()
	rb := NewRingBuffer(4, false)
	rb.Push("string")
	rb.Push(42)
	rb.Push(struct{ X int }{X: 7})
	rb.Push(nil) // RingBuffer accepts nil — Pop returns (nil, true)

	for _, want := range []interface{}{"string", 42, struct{ X int }{X: 7}, interface{}(nil)} {
		got, ok := rb.Pop()
		if !ok {
			t.Fatalf("Pop ok=false, expected ok=true for %v", want)
		}
		if got != want {
			t.Errorf("Pop = %v (%T), want %v (%T)", got, got, want, want)
		}
	}
}

// --- Existing benchmarks (preserved) ----------------------------------------

func BenchmarkChannel(b *testing.B) {
	str := "this is an input string\r\nand this is line 2"
	ch := make(chan string, 10)
	for i := 0; i < b.N; i++ {
		ch <- str
		_ = <-ch
	}
}

func BenchmarkRingBuffer(b *testing.B) {
	str := "this is an input string\r\nand this is line 2"
	rb := NewRingBuffer(10, false)
	for i := 0; i < b.N; i++ {
		_ = rb.Push(str)
		_, _ = rb.Pop()
	}
}
