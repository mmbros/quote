package taskengine

import (
	"math/rand"
	"sort"
)

// SortTasks reorder each worker tasks list,
// in order to have a worker handle each task as soon as possible.
//
// The algorithm is as follows.
//
// Until there are tasks left, loop over each worker and for each worker:
//
// 1. if it remains just one task in the worker list, take that
//
// 2. otherwise selects the task for which there are fewer workers left to assign it.
// Note that only the workers following the current one are considered
//
// 3. if multiple tasks meet the above criteria,
// select the task that has fewer workers that already have the task in the list
//
// 4. if multiple tasks meet the above criteria,
// select the task with the smallest TaskID.
// Note: used to make the algorithm deterministic.
func (wts WorkerTasks) SortTasks() {
	wtsCloned := wts.Clone()

	// WorkerTaks result
	res := WorkerTasks{}

	// The array of WorkerID
	wids := make([]WorkerID, 0, len(wtsCloned))

	// How many workers already have the TaskID assigned in the res object.
	doneMap := map[TaskID]int{}

	// Total number of tasks.
	// The same task is counted more times if assigned to different workers.
	// The loop below is repeated until this counter is zero.
	countTasks := 0

	for w, ts := range wtsCloned {
		res[w] = Tasks{}
		wids = append(wids, w)
		countTasks += len(ts)
	}

	// Sort the slice in order to be deterministic (for test pourposes)
	sort.Slice(wids, func(i, j int) bool {
		return wids[i] < wids[j]
	})

	for countTasks > 0 {

		// Inner loop over each worker
		for jw, wid := range wids {

			// Get the remaining tasks of the worker
			ts := wtsCloned[wid]
			if len(ts) == 0 {
				continue
			}

			var (
				minidx, mincount, mindone int
				mintid                    TaskID
			)
			first := true

			for idx, t := range ts {
				tid := t.TaskID()

				// count: how many workers, among the following workers,
				// can manage the task
				count := 0
				for j := jw + 1; j < len(wids); j++ {
					jwid := wids[j]
					for _, t := range wtsCloned[jwid] {
						if t.TaskID() == tid {
							count++
						}
					}
				}

				// done: how many workers already have the task in the list
				done := doneMap[tid]

				if first ||
					(count < mincount) ||
					(count == mincount && done < mindone) ||
					(count == mincount && done == mindone && tid < mintid) {
					mintid = tid
					minidx = idx
					mincount = count
					mindone = done
					first = false
				}
			}

			// updates variables
			doneMap[mintid] = mindone + 1
			res[wid] = append(res[wid], ts[minidx])
			countTasks--

			// remove the task from the worker tasks list
			L1 := len(ts) - 1
			ts[minidx] = ts[L1]
			ts = ts[:L1]
			wtsCloned[wid] = ts
		}
	}

	// update the current wts object with the result
	for wid := range wts {
		wts[wid] = res[wid]
	}

}

// SortRandom randomly reorder each worker tasks list,
func (wts WorkerTasks) SortRandom() {
	for w, ts := range wts {
		L := len(ts)
		permTs := make(Tasks, L)
		for i, j := range rand.Perm(L) {
			permTs[i] = ts[j]
		}
		wts[w] = permTs
	}
}
