package taskengine

import "sort"

// SortTasks reorder each worker tasks list, to handle globally each task as soon as possible.
func (wts WorkerTasks) SortTasks() {
	// obiettivo:
	//   per ogni worker ordinare i task in modo che ogni task sia globalmente gestito il prima possibile
	// ipotesi:
	//   tutti i task di un singolo worker sono distinti (TODO: VERIFICARE SE NECESSARIA)
	// strategia:
	//   ciclo sui worker. per ogni worker
	//   1. prendo il task con meno worker che l'hanno gia' assegnato
	//   2. a parita' del punto 1, prendo il task per cui rimangono meno worker
	//   3. a parita' del punto 2, prendo il task che ha meno worker gia' assegnati
	//   4. a parita' del punto 3, prendo il task con ID piu' piccolo (per rendere deterministico l'algoritmo)

	wtsCloned := wts.Clone()

	// initialize the WorkerTaks result and the array of wids
	res := WorkerTasks{}
	doneMap := map[TaskID]int{}
	wids := make([]WorkerID, 0, len(wtsCloned))
	countTasks := 0
	for w, ts := range wtsCloned {
		res[w] = Tasks{}
		wids = append(wids, w)
		countTasks += len(ts)
	}

	// sort the slice in order to be deterministic (for test pourposes)
	sort.Slice(wids, func(i, j int) bool {
		return wids[i] < wids[j]
	})

	for countTasks > 0 {

		for jw, wid := range wids {
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

				// count: conta fra i worker successivi quanti worker possono gestire il task
				count := 0
				for j := jw + 1; j < len(wids); j++ {
					jwid := wids[j]
					for _, t := range wtsCloned[jwid] {
						if t.TaskID() == tid {
							count++
						}
					}
				}
				// done: quanti worker hanno gia' il task nella lista
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

			// aggiorna le variabili
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

	// update wts each worker task lists
	for wid := range wts {
		wts[wid] = res[wid]
	}

}
