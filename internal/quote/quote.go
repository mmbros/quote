package quote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/mmbros/quote/internal/htmlquotescraper"
	"github.com/mmbros/quote/internal/htmlquotescraper/fondidocit"
	"github.com/mmbros/quote/internal/htmlquotescraper/fundsquarenet"
	"github.com/mmbros/quote/internal/htmlquotescraper/morningstarit"
	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/internal/quotegetterdb"
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
	Isin      string    `json:"isin,omitempty"`
	Source    string    `json:"source,omitempty"`
	Instance  int       `json:"instance"`
	URL       string    `json:"url,omitempty"`
	Price     float32   `json:"price,omitempty"`
	Currency  string    `json:"currency,omitempty"`
	Date      time.Time `json:"date,omitempty"`
	TimeStart time.Time `json:"time_start"`
	TimeEnd   time.Time `json:"time_end"`
	ErrMsg    string    `json:"error,omitempty"`
	Err       error     `json:"-"`
}

func (r *resultGetQuote) Success() bool {
	return r.Err == nil
}

func (r *resultGetQuote) dbInsert(db *quotegetterdb.QuoteDatabase) error {
	var qr *quotegetterdb.QuoteRecord

	// assert := func(b bool, label string) {
	// 	if !b {
	// 		panic("failed assert: " + label)
	// 	}
	// }

	// assert(r != nil, "r != nil")
	// assert(db != nil, "db != nil")

	// skip context.Canceled errors
	if r.Err != nil {
		if err, ok := r.Err.(*htmlquotescraper.Error); ok {
			if !errors.Is(err, context.Canceled) {
				return nil
			}
		}
	}
	qr = &quotegetterdb.QuoteRecord{
		Isin:     r.Isin,
		Source:   r.Source,
		Price:    r.Price,
		Currency: r.Currency,
		Date:     r.Date,
		URL:      r.URL,
		ErrMsg:   r.ErrMsg,
	}
	// isin and source are mandatory
	// assert(len(qr.Isin) > 0, "len(qr.Isin) > 0")
	// assert(len(qr.Source) > 0, "len(qr.Source) > 0")

	// save to database
	return db.InsertQuotes(qr)
}

func dbInsert(dbpath string, results []*resultGetQuote) error {
	if len(dbpath) == 0 {
		return nil
	}

	// save to database
	db, err := quotegetterdb.Open(dbpath)
	if db != nil {
		defer db.Close()

		for _, r := range results {
			err = r.dbInsert(db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Get is ..
func Get(isins []string, sources []string, workers []int, dbpath string) error {

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
				Instance:  inst,
				TimeStart: time1,
				TimeEnd:   time2,
				Err:       err,
			}
			if res != nil {
				r.Isin = res.Isin
				r.Source = res.Name
				r.Date = res.Date
				r.Price = res.Price
				r.Currency = res.Currency
			}
			if err != nil {
				r.ErrMsg = err.Error()
				if e, ok := err.(*htmlquotescraper.Error); ok {
					r.Isin = e.Isin
					r.Source = e.Name
					r.URL = e.URL
				}
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

	resChan, err := taskengine.Execute(context.Background(), ws, wts, taskengine.FirstSuccessThenCancel)
	if err != nil {
		return err
	}

	results := []*resultGetQuote{}
	for r := range resChan {
		res := r.(*resultGetQuote)
		results = append(results, res)
	}

	// save to database, if not empty
	err = dbInsert(dbpath, results)
	if err != nil {
		fmt.Println(err)
	}

	json, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(json))

	return nil
}
