package quote

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/mmbros/quote/internal/htmlquotescraper/fondidocit"
	"github.com/mmbros/quote/internal/htmlquotescraper/fundsquarenet"
	"github.com/mmbros/quote/internal/htmlquotescraper/morningstarit"
	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/pkg/taskengine"
)

var (
	quoteGetter = make(map[string]quotegetter.QuoteGetter)
)

func init() {
	type fnNewQuoteGetter func(string) quotegetter.QuoteGetter

	src := map[string]fnNewQuoteGetter{
		"fondidocit":    fondidocit.NewQuoteGetter,
		"morningstarit": morningstarit.NewQuoteGetter,
		"fundsquarenet": fundsquarenet.NewQuoteGetter,
	}

	for name, fn := range src {
		quoteGetter[name] = fn(name)
	}

}

// getSources returns a list of the names of the available quoteGetters.
func getSources() []string {

	list := make([]string, 0, len(quoteGetter))
	for name := range quoteGetter {
		list = append(list, name)
	}

	return list
}

// Sources returns a sorted list of the names of the avaliable quoteGetters.
func Sources() []string {
	list := getSources()
	sort.Strings(list)
	return list
}

// getFilteresSources verified the passed sources names.
// It returns nil if are all available, an error otherwise.
func getFilteredSources(names []string) ([]string, error) {
	if len(names) == 0 {
		return getSources(), nil
	}
	for _, name := range names {
		if _, ok := quoteGetter[name]; !ok {
			return nil, fmt.Errorf("source not available: %q", name)
		}
	}
	return names, nil
}

type taskGetQuote struct {
	isin string
	url  string
}

func (t *taskGetQuote) TaskID() taskengine.TaskID {
	return taskengine.TaskID(t.isin)
}

type resultGetQuote struct {
	*quotegetter.Result
	ScraperInst int
	TimeStart   time.Time
	TimeEnd     time.Time
	Err         error
}

func (r *resultGetQuote) Success() bool {
	return r.Err == nil
}

// // GetOld is ..
// func GetOld(isins []string, sources []string, workers int) error {

// 	// array of the used sources
// 	filteredSources, err := getFilteredSources(sources)
// 	if err != nil {
// 		return err
// 	}

// 	// Tasks
// 	ts := make(taskengine.Tasks, 0, len(isins))
// 	for _, isin := range isins {
// 		ts = append(ts, &taskGetQuote{isin, ""})
// 	}

// 	// Workers
// 	ws := make([]*taskengine.Worker, 0, len(filteredSources))

// 	// WorkerTasks
// 	wts := make(taskengine.WorkerTasks)

// 	for _, name := range filteredSources {

// 		qg := quoteGetter[name]

// 		// work function for the named source
// 		wfn := func(ctx context.Context, inst int, task taskengine.Task) taskengine.Result {
// 			t := task.(*taskGetQuote)
// 			time1 := time.Now()
// 			res, err := qg.GetQuote(ctx, t.isin, t.url)
// 			time2 := time.Now()

// 			r := &resultGetQuote{
// 				Result:      res,
// 				ScraperInst: inst,
// 				TimeStart:   time1,
// 				TimeEnd:     time2,
// 				Err:         err,
// 			}
// 			return r
// 		}

// 		w := &taskengine.Worker{
// 			WorkerID:  taskengine.WorkerID(name),
// 			Instances: workers,
// 			Work:      wfn,
// 		}
// 		ws = append(ws, w)

// 		// set the same tasks for all the workers
// 		wts[w.WorkerID] = ts
// 	}

// 	wts.SortTasks()

// 	resChan, err := taskengine.Execute(context.Background(), ws, wts)
// 	if err != nil {
// 		return err
// 	}

// 	results := []*resultGetQuote{}
// 	for r := range resChan {
// 		res := r.(*resultGetQuote)
// 		results = append(results, res)
// 	}

// 	json, err := json.MarshalIndent(results, "", " ")
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(string(json))

// 	return nil
// }

// Get is ..
func Get(isins []string, sources []string, workers []int) error {

	lenWorkers1 := len(workers) - 1
	getSourceWorkers := func(idx int) int {
		if idx > lenWorkers1 {
			idx = lenWorkers1
		}
		if idx < 0 {
			return 1 // default number of instances
		}
		return workers[idx]
	}

	// array of the used sources
	filteredSources, err := getFilteredSources(sources)
	if err != nil {
		return err
	}

	// Tasks
	ts := make(taskengine.Tasks, 0, len(isins))
	for _, isin := range isins {
		ts = append(ts, &taskGetQuote{isin, ""})
	}

	// Workers
	ws := make([]*taskengine.Worker, 0, len(filteredSources))

	// WorkerTasks
	wts := make(taskengine.WorkerTasks)

	for srcIdx, srcName := range filteredSources {

		qg := quoteGetter[srcName]

		// work function for the named source
		wfn := func(ctx context.Context, inst int, task taskengine.Task) taskengine.Result {
			t := task.(*taskGetQuote)
			time1 := time.Now()
			res, err := qg.GetQuote(ctx, t.isin, t.url)
			time2 := time.Now()

			r := &resultGetQuote{
				Result:      res,
				ScraperInst: inst,
				TimeStart:   time1,
				TimeEnd:     time2,
				Err:         err,
			}
			return r
		}

		w := &taskengine.Worker{
			WorkerID:  taskengine.WorkerID(srcName),
			Instances: getSourceWorkers(srcIdx),
			Work:      wfn,
		}
		ws = append(ws, w)

		// set the same tasks for all the workers
		wts[w.WorkerID] = ts
	}

	wts.SortTasks()

	resChan, err := taskengine.Execute(context.Background(), ws, wts)
	if err != nil {
		return err
	}

	results := []*resultGetQuote{}
	for r := range resChan {
		res := r.(*resultGetQuote)
		results = append(results, res)
	}

	json, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(json))

	return nil
}
