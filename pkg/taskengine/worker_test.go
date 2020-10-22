package taskengine

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testCaseTask struct {
	taskid  string
	msec    int
	success bool
}
type testTask struct {
	testCaseTask
	workerid string
	t        *testing.T
}
type testResult struct {
	taskid     string
	workerid   string
	workerinst int
	result     string
	err        error
}

type simpleResult struct {
	taskid   string
	workerid string
	success  bool
}

func (t *testCaseTask) TaskID() TaskID                 { return TaskID(t.taskid) }
func (t *testCaseTask) Equal(other *testCaseTask) bool { return t.taskid == other.taskid }
func (t *testCaseTask) String() string                 { return string(t.taskid) }

func (res *testResult) Success() bool { return res.err == nil }
func (res *testResult) Status() string {
	if res.err == nil {
		return "SUCCESS"
	}
	return res.err.Error()
}
func (res *testResult) Simple() *simpleResult {
	return &simpleResult{
		taskid:   res.taskid,
		workerid: res.workerid,
		success:  res.Success(),
	}
}

func (res *simpleResult) String() string {
	return fmt.Sprintf("{tid:%q, wid:%q, success:%v}", res.taskid, res.workerid, res.success)
}

func (res *simpleResult) Equal(other *simpleResult) bool {
	return res.taskid == other.taskid &&
		res.workerid == other.workerid &&
		res.success == other.success
}

// simpleResults implements sort.Interface for []*simpleResult
type simpleResults []*simpleResult

func (a simpleResults) Len() int      { return len(a) }
func (a simpleResults) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a simpleResults) Less(i, j int) bool {
	return (a[i].taskid < a[j].taskid) ||
		(a[i].taskid == a[j].taskid && a[i].workerid < a[j].workerid)
}

// newTestWorkeridTasks creates a WorkerTasks object from a map workerId -> [taskId1, taskId2, ...]
func newTestWorkeridTasks(t *testing.T, wts map[string][]testCaseTask) WorkerTasks {
	wtasks := WorkerTasks{}

	for wid, tcts := range wts {
		ts := Tasks{}
		for _, tct := range tcts {
			tt := &testTask{tct, wid, t}
			if tct.msec <= 0 {
				tt.msec = 10
			}
			ts = append(ts, tt)
		}
		wtasks[WorkerID(wid)] = ts
	}
	return wtasks
}

func workFn(ctx context.Context, workerInst int, task Task) Result {

	ttask := task.(*testTask)
	if ttask == nil {
		panic("task is not a testTask: ahhh")
	}

	tres := &testResult{
		taskid:     ttask.taskid,
		workerid:   ttask.workerid,
		workerinst: workerInst,
		result:     fmt.Sprintf("%dms", ttask.msec),
	}

	ttask.t.Logf("WORKING:   (%s, %s)", ttask.workerid, ttask.taskid)

	select {
	case <-ctx.Done():
		tres.err = ctx.Err()
	case <-time.After(time.Duration(ttask.msec) * time.Millisecond):
		if !ttask.success {
			tres.err = errors.New("ERR")
		}
	}

	ttask.t.Logf("WORKED: (%s, %s) -> %s", ttask.workerid, ttask.taskid, tres.Status())

	return Result(tres)
}

/*
func TestTasksString(t *testing.T) {
	tasks := Tasks{
		&testCaseTask{"t1", 0, true},
		&testCaseTask{"t3", 0, true},
		&testCaseTask{"t2", 0, true},
	}

	expected := "[t1, t3, t2]"
	found := tasks.String()

	if found != expected {
		t.Errorf("Expected %s, found %s", expected, found)
	}

}
*/
func TestExecuteFirstSuccessOrLastError(t *testing.T) {
	workers := []*Worker{
		{"w1", 1, workFn},
		{"w2", 1, workFn},
		{"w3", 1, workFn},
	}

	type testCase struct {
		input    map[string][]testCaseTask
		expected simpleResults
	}

	testCases := map[string]testCase{
		"all ok": {
			input: map[string][]testCaseTask{
				"w1": {{"t3", 30, true}, {"t2", 20, true}, {"t1", 10, true}},
				"w2": {{"t3", 20, true}, {"t2", 10, true}},
				"w3": {{"t3", 10, true}},
			},
			expected: simpleResults{
				{"t1", "w1", true},
				{"t2", "w2", true},
				{"t3", "w3", true},
			},
		},
		"w3-t3 ko": {
			input: map[string][]testCaseTask{
				"w1": {{"t3", 30, true}, {"t2", 20, true}, {"t1", 10, true}},
				"w2": {{"t3", 20, true}, {"t2", 10, true}},
				"w3": {{"t3", 10, false}},
			},
			expected: simpleResults{
				{"t1", "w1", true},
				{"t2", "w2", true},
				{"t3", "w2", true},
			},
		},
		"all ko": {
			input: map[string][]testCaseTask{
				"w1": {{"t3", 30, false}, {"t2", 20, false}, {"t1", 10, false}},
				"w2": {{"t3", 20, false}, {"t2", 10, false}},
				"w3": {{"t3", 10, false}},
			},
			expected: simpleResults{
				{"t1", "w1", false},
				{"t2", "w1", false},
				{"t3", "w1", false},
			},
		},
		"all ok w1": {
			input: map[string][]testCaseTask{
				"w1": {{"t1", 10, true}, {"t2", 10, true}, {"t3", 10, true}},
				"w2": {{"t2", 40, true}, {"t3", 20, true}},
				"w3": {{"t3", 50, true}},
			},
			expected: simpleResults{
				{"t1", "w1", true},
				{"t2", "w1", true},
				{"t3", "w1", true},
			},
		},
		"all ok w1 w2": {
			input: map[string][]testCaseTask{
				"w1": {{"t1", 10, true}, {"t2", 10, true}, {"t3", 20, true}},
				"w2": {{"t2", 50, true}, {"t3", 10, true}},
				"w3": {{"t3", 50, true}},
			},
			expected: simpleResults{
				{"t1", "w1", true},
				{"t2", "w1", true},
				{"t3", "w2", true},
			},
		},
		"all ko w1 but t1": {
			input: map[string][]testCaseTask{
				"w1": {{"t1", 10, true}, {"t2", 10, false}, {"t3", 10, false}},
				"w2": {{"t3", 30, false}, {"t2", 10, true}},
				"w3": {{"t3", 50, true}},
			},
			expected: simpleResults{
				{"t1", "w1", true},
				{"t2", "w2", true},
				{"t3", "w3", true},
			},
		},
	}
	// copts := cmp.Options{
	// 	// cmpopts.IgnoreUnexported(testCaseTask{}),
	// 	cmpopts.SortSlices(func(a, b simpleResult) bool {
	// 		return (a.taskid < b.taskid) ||
	// 			(a.taskid == b.taskid && a.workerid < b.workerid)
	// 	}),
	// }
	for title, tc := range testCases {
		tasks := newTestWorkeridTasks(t, tc.input)
		ctx := context.Background()
		out, err := Execute(ctx, workers, tasks, FirstSuccessOrLastError)
		if err != nil {
			t.Fatal(err.Error())
		}

		results := simpleResults{}

		for res := range out {
			tres := res.(*testResult)
			results = append(results, tres.Simple())
			// t.Logf("RESULT: (%s, %s) -> %s - %s", tres.taskid, tres.workerid, tres.result, tres.Status())

		}

		// if diff := cmp.Diff(tc.expected, results, copts); diff != "" {
		// 	t.Errorf("%s: mismatch (-want +got):\n%s", title, diff)
		// }
		sort.Sort(results)
		sort.Sort(tc.expected)
		if results.Len() != tc.expected.Len() {
			t.Errorf("%s - Expected %d, got %d results", title, tc.expected.Len(), results.Len())
			t.Logf("expected: %v", tc.expected)
			t.Logf("got: %v", results)
			t.FailNow()
		} else {
			for idx, res := range results {
				exp := tc.expected[idx]
				if !res.Equal(exp) {
					t.Errorf("%s - Expected %v, found %v", title, exp, res)
				}
			}
		}

	}
	// t.FailNow()
}

func TestExecuteFirstSuccessThenCancel(t *testing.T) {
	workers := []*Worker{
		{"w1", 1, workFn},
		{"w2", 1, workFn},
		{"w3", 1, workFn},
	}

	type testCase struct {
		input    map[string][]testCaseTask
		expected simpleResults
	}

	testCases := map[string]testCase{
		"all ok": {
			input: map[string][]testCaseTask{
				"w1": {{"t1", 300, true}},
				"w2": {{"t1", 200, true}},
				"w3": {{"t1", 100, true}},
			},
			expected: simpleResults{
				{"t1", "w3", true},
				{"t1", "w2", false},
				{"t1", "w1", false},
			},
		},
		"two ok": {
			input: map[string][]testCaseTask{
				"w1": {{"t1", 300, true}},
				"w2": {{"t1", 200, true}},
				"w3": {{"t1", 100, false}},
			},
			expected: simpleResults{
				{"t1", "w3", false},
				{"t1", "w2", true},
				{"t1", "w1", false},
			},
		},
	}

	// NOTE: it is important to use *simpleResult and not simpleResult
	lessFunc := func(a, b *simpleResult) bool {
		return (a.taskid < b.taskid) ||
			(a.taskid == b.taskid && a.workerid < b.workerid) ||
			(a.taskid == b.taskid && a.workerid == b.workerid && !a.success)
	}

	for title, tc := range testCases {

		tasks := newTestWorkeridTasks(t, tc.input)
		ctx := context.Background()
		out, err := Execute(ctx, workers, tasks, FirstSuccessThenCancel)
		if err != nil {
			t.Fatal(err.Error())
		}

		results := simpleResults{}

		for res := range out {
			tres := res.(*testResult)
			results = append(results, tres.Simple())
			// t.Logf("RESULT: (%s, %s) -> %s - %s", tres.taskid, tres.workerid, tres.result, tres.Status())

		}

		copts := cmp.Options{
			cmpopts.SortSlices(lessFunc),
		}
		if diff := cmp.Diff(tc.expected, results, copts); diff != "" {
			t.Errorf("%s: mismatch (-want +got):\n%s", title, diff)
		}
	}
}

func TestExecuteAll(t *testing.T) {
	workers := []*Worker{
		{"w1", 1, workFn},
		{"w2", 1, workFn},
		{"w3", 1, workFn},
	}

	type testCase struct {
		input    map[string][]testCaseTask
		expected simpleResults
	}

	testCases := map[string]testCase{
		"all ok": {
			input: map[string][]testCaseTask{
				"w1": {{"t1", 300, true}},
				"w2": {{"t1", 200, true}},
				"w3": {{"t1", 100, true}},
			},
			expected: simpleResults{
				{"t1", "w3", true},
				{"t1", "w2", true},
				{"t1", "w1", true},
			},
		},
	}

	for title, tc := range testCases {

		tasks := newTestWorkeridTasks(t, tc.input)
		ctx := context.Background()
		out, err := Execute(ctx, workers, tasks, All)
		if err != nil {
			t.Fatal(err.Error())
		}

		results := simpleResults{}

		for res := range out {
			tres := res.(*testResult)
			results = append(results, tres.Simple())
			// t.Logf("RESULT: (%s, %s) -> %s - %s", tres.taskid, tres.workerid, tres.result, tres.Status())

		}

		if diff := cmp.Diff(tc.expected, results, nil); diff != "" {
			t.Errorf("%s: mismatch (-want +got):\n%s", title, diff)
		}
	}
}
