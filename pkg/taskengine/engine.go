package taskengine

import (
	"context"
	"fmt"
	"sync"
)

// Engine is ...
type Engine struct {
	workers map[WorkerID]*Worker
	tasks   WorkerTasks // map[WorkerID]*Tasks
	tidctxs map[TaskID]*taskIDContext
}

// taskIDContext contains the information common to all the task with the same TaskID.
// Note: the Task itself cannot be here, because different task with the same TaskID
//       can have different information.
type taskIDContext struct {
	taskID TaskID

	// number of workers that can handle the TaskID
	workers int

	// TaskID context
	ctx context.Context

	// TaskID context cancel function
	cancel context.CancelFunc

	// result channel of the TaskID
	resChan chan Result
}

type workerRequest struct {
	ctx context.Context

	// task of the worker
	task Task

	// response channel for the specific taskid
	resChan chan Result
}

// NewEngine initialize a new engine object from the list of workers and the tasks of each worker.
// It performs some sanity check and return error in case of incongruences.
func NewEngine(ctx context.Context, ws []*Worker, wts WorkerTasks) (*Engine, error) {

	if ctx == nil {
		return nil, fmt.Errorf("Nil context")
	}

	// check workers and build a map from workerid to Worker
	workers := map[WorkerID]*Worker{}
	for _, w := range ws {
		if _, ok := workers[w.WorkerID]; ok {
			return nil, fmt.Errorf("Duplicate worker: WorkerID=%q", w.WorkerID)
		}
		if w.Instances <= 0 || w.Instances > maxInstances {
			return nil, fmt.Errorf("Instances must be in 1..%d range: WorkerID=%q", maxInstances, w.WorkerID)
		}
		if w.Work == nil {
			return nil, fmt.Errorf("Work function cannot be nil: WorkerID=%q", w.WorkerID)
		}
		workers[w.WorkerID] = w
	}

	// create each taskID context
	tidctxs := map[TaskID]*taskIDContext{}
	tasks := WorkerTasks{}

	for wid, ts := range wts {

		// for not empty task lists, check the worker exists
		if len(ts) > 0 {
			if _, ok := workers[wid]; !ok {
				return nil, fmt.Errorf("Tasks for undefined worker: WorkerID=%q", wid)
			}
			// save the task list of the worker in the engine
			tasks[wid] = ts
		}

		// create a taskIDContext object for each different TaskID
		for _, t := range ts {
			tid := t.TaskID()
			tidctx := tidctxs[tid]
			if tidctx == nil {
				// new TaskID found: create a new context for the task
				tidctx = &taskIDContext{taskID: tid}

				// save the context to the map
				tidctxs[tid] = tidctx

				// NOTE: in case of buffered chan, we can't create here the resChan
				//       because we don't know yet the number of workers
				//       that will handle the specific task
				//       -> the buffered chan will be created after the loop
				//       Also context and cancel function will be created after the loop
			}
			// Increment the number of workers that handle the task
			// NOTE: doesn't check if the worker has already been used for the same task
			tidctx.workers++
		}
	}

	// complete the creation of the taskIDContext
	for _, tidctx := range tidctxs {
		// create the context and cancel function
		tidctx.ctx, tidctx.cancel = context.WithCancel(ctx)

		// create the resChan buffered channel
		tidctx.resChan = make(chan Result, tidctx.workers)
	}

	return &Engine{workers, tasks, tidctxs}, nil
}

// createWorkerRequestChan returns a chan where are enqueued the worker's requests
func (eng *Engine) createWorkerRequestChan(wid WorkerID) chan *workerRequest {
	out := make(chan *workerRequest)
	go func() {
		// loop for each task of the worker
		for _, t := range eng.tasks[wid] {
			tidctx := eng.tidctxs[t.TaskID()]

			req := &workerRequest{
				ctx:     tidctx.ctx,
				resChan: tidctx.resChan,
				task:    t,
			}
			out <- req
		}
		close(out)
	}()
	return out
}

// getFirstSuccessResultIfAny send to the out channel a single result for the job task.
// It is the first success response or the last response (success or error).
func (tidctx *taskIDContext) getFirstSuccessResultIfAny(out chan Result) {
	todo := true
	count := tidctx.workers

	for ; count > 0; count-- {

		select {
		case res := <-tidctx.resChan:
			// if not already done,
			// send the result if Success,
			// or if it is the last result.

			fmt.Print(res)

			if todo && (res.Success() || count == 1) {
				todo = false
				tidctx.cancel()
				out <- res
				// XXX: inserted return statement: check it!!!
				//return
			}
		case <-tidctx.ctx.Done():
			tidctx.cancel()
		}
	}
}

// getAllFirstSuccessResultIfAny send to the out channel a single result for the job task.
// It is the first success response or the last response (success or error).
func (tidctx *taskIDContext) getAllFirstSuccessResultIfAny(out chan Result) {
	todo := true
	count := tidctx.workers

	for ; count > 0; count-- {

		res := <-tidctx.resChan
		// if Success and not already done, cancel the context
		if todo && res.Success() {
			todo = false
			tidctx.cancel()
		}
		out <- res
	}
}

// // getAllFirstSuccessResultIfAny send to the out channel a single result for the job task.
// // It is the first success response or the last response (success or error).
// func (tidctx *taskIDContext) DONOTWORK_getAllFirstSuccessResultIfAny(out chan Result) {
// 	todo := true

// 	// it doesn't work, because resChan it is never closed, I suppose...
// 	for res := range tidctx.resChan {
// 		if todo && res.Success() {
// 			todo = false
// 			tidctx.cancel()
// 		}
// 		out <- res
// 	}
// }

func (tidctx *taskIDContext) getAllResults(out chan Result) {
	count := tidctx.workers

	for ; count > 0; count-- {

		select {
		case res := <-tidctx.resChan:
			out <- res
		case <-tidctx.ctx.Done():
			tidctx.cancel()
		}
	}
}

// Execute returns a chan that receives the Results of the workers for the input Requests.
func (eng *Engine) Execute() (chan Result, error) {

	if eng == nil {
		return nil, fmt.Errorf("Engine is nil")
	}

	// Creates the output channel
	out := make(chan Result)

	// Starts a goroutine for each different TaskID to wait for the result
	var wg sync.WaitGroup
	wg.Add(len(eng.tidctxs))
	for _, t := range eng.tidctxs {
		go func(tidctx *taskIDContext) {
			// tidctx.getFirstSuccessResultIfAny(out)
			// tidctx.getAllResults(out)
			tidctx.getAllFirstSuccessResultIfAny(out)
			wg.Done()
		}(t)
	}

	// Start a goroutine to close the out channel once all the output
	// goroutines are done. This must start after the wg.Add call.
	go func() {
		wg.Wait()
		//log.Println("CLOSING OUT")
		close(out)
	}()

	// Starts the goroutines that executes the real work.
	// For each worker it starts N goroutines, with N = Instances.
	// Each goroutine get the input from the worker request channel,
	// and put the output to the task result channel (contained in the request).
	for wid, worker := range eng.workers {

		// create the worker request channel of the worker
		wreqChan := eng.createWorkerRequestChan(wid)

		// for each worker instances
		for i := 0; i < worker.Instances; i++ {

			go func(w *Worker, workerInst int, reqc <-chan *workerRequest) {
				for req := range reqc {
					// send the worker result of the task,
					// to the response chan of the task
					req.resChan <- w.Work(req.ctx, workerInst, req.task)
				}
			}(worker, i, wreqChan)
		}
	}

	return out, nil

}
