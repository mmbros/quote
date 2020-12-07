package taskengine

import (
	"context"
)

// Max number of instances for each worker
const maxInstances = 100

//-----------------------------------------------------------------------------
// Types to be customized if needed. For example:
//   type TaskID    int
//   type WorkerID  int

// WorkerID type definition.
type WorkerID string

// TaskID type definition.
type TaskID string

//-----------------------------------------------------------------------------

// Task is a unit of work that can be executed by a worker
// Two or more task with the same TaskID are equivalent
// and possibly only one will be executed.
// Two or more task with the same TaskID can contain different information
// usefull for a specific  worker.
type Task interface {
	TaskID() TaskID
}

// Result is the interface that must be matched by the output of the Work function.
type Result interface {
	// Success return true in case of a success response.
	// In this case no other Request will be worked for the same Job.
	Success() bool
}

// WorkFunc is the worker function.
// - context: the context
// - int:     the instance id of the worker
// - Task:    the task to be eecuted
type WorkFunc func(context.Context, int, Task) Result

// Worker is the unit (identified by WorkerID)
// that receives the Requests and
// executes a specific WorkFunc function to return the Responses.
// The Instances parameters represents the number of instances of each worker
type Worker struct {

	// Unique ID of the worker
	WorkerID WorkerID

	// Number of worker instances. Must be greater or equal 1
	Instances int

	// The work function
	Work WorkFunc
}

// Tasks is an array of tasks.
type Tasks []Task

// WorkerTasks is a map representing the tasks list of each worker
type WorkerTasks map[WorkerID]Tasks

// Clone method returns a cloned copy of the WorkerTasks object.
func (wts WorkerTasks) Clone() WorkerTasks {
	wts2 := WorkerTasks{}
	for w, ts := range wts {
		ts2 := Tasks{}
		for _, t := range ts {
			ts2 = append(ts2, t)
		}
		wts2[w] = ts2
	}
	return wts2
}

// // String representation of a Tasks object.
// func (ts Tasks) String() string {
// 	a := make([]string, 0, len(ts))
// 	for _, t := range ts {
// 		a = append(a, string(t.TaskID()))
// 	}
// 	return "[" + strings.Join(a, ", ") + "]"
// }

// Remove removes the i-th task of the list.
// It returns the rask removed.
func (ts *Tasks) Remove(i int) Task {
	t := (*ts)[i]
	L1 := len(*ts) - 1
	(*ts)[i] = (*ts)[L1]
	(*ts) = (*ts)[:L1]
	return t
}

// // String representation of a WorkerTasks object.
// func (wts WorkerTasks) String() string {

// 	// build the array of (the string representation of) workerID sorted alphabetically
// 	wids := make([]string, 0, len(wts))
// 	for wid := range wts {
// 		wids = append(wids, string(wid))
// 	}
// 	sort.Strings(wids)

// 	// print to a buffer
// 	buf := new(bytes.Buffer)
// 	fmt.Fprintf(buf, "{\n")
// 	for _, wid := range wids {
// 		fmt.Fprintf(buf, "   %s : %s\n", wid, wts[WorkerID(wid)])
// 	}
// 	fmt.Fprintf(buf, "}\n")

// 	// buffer to string
// 	return buf.String()
// }

// Execute function returns a chan that receives the Results of the workers for the input Requests.
func Execute(ctx context.Context, workers []*Worker, tasks WorkerTasks, mode Mode) (chan Result, error) {
	eng, err := newEngine(ctx, workers, tasks)
	if err != nil {
		return nil, err
	}

	return eng.Execute(mode)
}
