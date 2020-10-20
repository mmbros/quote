package taskengine

import (
	"context"
	"errors"
	"testing"
)

func TestNewEngineNilContext(t *testing.T) {

	workers := []*Worker{
		{"w1", 1, workFn},
		{"w2", 2, workFn},
	}
	input := map[string][]testCaseTask{
		"w1": {{"t3", 30, true}, {"t2", 20, true}, {"t1", 10, true}},
		"w2": {{"t2", 10, true}},
	}

	tasks := newTestWorkeridTasks(t, input)
	_, err := NewEngine(nil, workers, tasks)
	expected := errors.New("Nil context")

	if err == nil {
		t.Errorf("Expected error %q, found no error", expected)
	} else if err.Error() != expected.Error() {
		t.Errorf("Expected error %q, found error %q", expected, err)
	}
}

func TestNewEngine(t *testing.T) {

	type testCase struct {
		workers []*Worker
		input   map[string][]testCaseTask
		err     error
	}

	testCases := map[string]testCase{
		"duplicate worker": {
			workers: []*Worker{
				{"w1", 1, workFn},
				{"w2", 2, workFn},
				{"w1", 3, workFn},
			},
			input: map[string][]testCaseTask{},
			err:   errors.New("Duplicate worker: WorkerID=\"w1\""),
		},
		"instances < 1": {
			workers: []*Worker{
				{"w1", 1, workFn},
				{"w2", 2, workFn},
				{"w3", 0, workFn},
			},
			input: map[string][]testCaseTask{},
			err:   errors.New("Instances must be in 1..100 range: WorkerID=\"w3\""),
		},
		"instances > 100": {
			workers: []*Worker{
				{"w1", 1, workFn},
				{"w2", 2, workFn},
				{"w3", 101, workFn},
			},
			input: map[string][]testCaseTask{},
			err:   errors.New("Instances must be in 1..100 range: WorkerID=\"w3\""),
		},
		"ko work function": {
			workers: []*Worker{
				{"w1", 1, workFn},
				{"w2", 2, nil},
				{"w3", 3, workFn},
			},
			input: map[string][]testCaseTask{},
			err:   errors.New("Work function cannot be nil: WorkerID=\"w2\""),
		},
		"undefined worker": {
			workers: []*Worker{
				{"w1", 1, workFn},
				{"w2", 2, workFn},
				{"w3", 3, workFn},
			},
			input: map[string][]testCaseTask{
				"w1":   {{"t3", 30, true}, {"t2", 20, true}, {"t1", 10, true}},
				"w000": {{"t3", 10, true}},
				"w2":   {{"t3", 20, true}, {"t2", 10, true}},
			},
			err: errors.New("Tasks for undefined worker: WorkerID=\"w000\""),
		},
	}

	ctx := context.Background()
	for title, tc := range testCases {
		tasks := newTestWorkeridTasks(t, tc.input)
		_, err := NewEngine(ctx, tc.workers, tasks)

		if tc.err == nil {
			if err != nil {
				t.Errorf("%s - Unexpected error %q", title, err)
			}
		} else {
			// tc.err != nil
			if err == nil {
				t.Errorf("%s - Expected error %q, found no error", title, tc.err)
			} else if err.Error() != tc.err.Error() {
				t.Errorf("%s - Expected error %q, found error %q", title, tc.err, err)
			}
		}
	}
}
