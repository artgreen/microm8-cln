package core

import (
	"fmt"
	"testing"
	"time"

	"paleotronic.com/log"
)

type ThingWhatDoes struct {
	TaskPerformer
	running bool
}

func (th *ThingWhatDoes) handleTask(task *Task) (interface{}, error) {
	log.Printf("Got a task: action = %s", task.Action)
	switch task.Action {
	case "stop":
		th.running = false
		return "ok", nil
	}
	return "", fmt.Errorf("unrecognised task: %s", task.Action)
}

func (th *ThingWhatDoes) Run(t *testing.T) {

	go func() {
		th.running = true
		for th.running {
			t.Log("running")
			time.Sleep(1 * time.Second)
			th.HandleTasks(th.handleTask)
		}
	}()

}

// TestTaskExecution is an unfinished scaffolding test (pre-Phase-1).
// It currently has a `t.Fail()` at the end without an associated
// condition, AND a goroutine-vs-test race because thing.Run(t) writes
// to thing.running while the main test reads it. Skipping until the
// test is rewritten with proper synchronisation.
func TestTaskExecution(t *testing.T) {
	t.Skip("TaskPerformer test needs rewrite: t.Fail() at end + concurrent " +
		"access to thing.running without synchronisation")

	thing := &ThingWhatDoes{}
	go thing.Run(t)

	tt := NewTask("stop")
	resp := tt.Request(thing)

	if resp.Err != nil {
		t.Fatal("Expected resp.Err to be nil")
	}

	if thing.running {
		t.Fatal("Expected thing to stop running")
	}
}
