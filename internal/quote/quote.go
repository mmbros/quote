package quote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/internal/quotegetter/cryptonatorcom"
	"github.com/mmbros/quote/internal/quotegetter/scrapers"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/fondidocit"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/fundsquarenet"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/morningstarit"
	"github.com/mmbros/quote/internal/quotegetterdb"
	"github.com/mmbros/quote/pkg/taskengine"
)

// SourceIsins struct represents the isins to get from a specific source
type SourceIsins struct {
	Source  string   `json:"source,omitempty"`
	Workers int      `json:"workers,omitempty"`
	Proxy   string   `json:"proxy,omitempty"`
	Isins   []string `json:"isins,omitempty"`
}

var (
	quoteGetter = make(map[string]quotegetter.QuoteGetter)
)

func init() {
	type fnNewQuoteGetter func(string) quotegetter.QuoteGetter

	fnCryptonatorcom := func(currency string) fnNewQuoteGetter {
		return func(name string) quotegetter.QuoteGetter {
			return cryptonatorcom.NewQuoteGetter(name, currency)
		}
	}

	src := map[string]fnNewQuoteGetter{
		"fondidocit":         fondidocit.NewQuoteGetter,
		"morningstarit":      morningstarit.NewQuoteGetter,
		"fundsquarenet":      fundsquarenet.NewQuoteGetter,
		"cryptonatorcom-EUR": fnCryptonatorcom("EUR"),
		"cryptonatorcom-USD": fnCryptonatorcom("USD"),
	}

	for name, fn := range src {
		qg := fn(name)
		quoteGetter[qg.Name()] = qg
	}

}

type taskGetQuote struct {
	isin string
	url  string
	// proxy string
}

func (t *taskGetQuote) TaskID() taskengine.TaskID {
	return taskengine.TaskID(t.isin)
}

// resultGetQuote.Date field is a pointer in order to omit zero dates.
// see https://stackoverflow.com/questions/32643815/json-omitempty-with-time-time-field

type resultGetQuote struct {
	Isin      string     `json:"isin,omitempty"`
	Source    string     `json:"source,omitempty"`
	Instance  int        `json:"instance"`
	URL       string     `json:"url,omitempty"`
	Price     float32    `json:"price,omitempty"`
	Currency  string     `json:"currency,omitempty"`
	Date      *time.Time `json:"date,omitempty"` // need a pointer to omit zero date
	TimeStart time.Time  `json:"time_start"`
	TimeEnd   time.Time  `json:"time_end"`
	ErrMsg    string     `json:"error,omitempty"`
	Err       error      `json:"-"`
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
		if err, ok := r.Err.(*scrapers.Error); ok {
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
		URL:      r.URL,
		ErrMsg:   r.ErrMsg,
	}
	if r.Date != nil {
		qr.Date = *r.Date
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

func checkListOfSourceIsins(items []SourceIsins) error {
	used := map[string]struct{}{}

	for _, item := range items {

		if _, ok := used[item.Source]; ok {
			return fmt.Errorf("duplicate source %q", item.Source)
		}
		used[item.Source] = struct{}{}

		if _, ok := quoteGetter[item.Source]; !ok {
			return fmt.Errorf("source %q not available", item.Source)
		}
		if item.Workers <= 0 {
			return fmt.Errorf("source %q with invalid workers %d", item.Source, item.Workers)
		}
	}
	return nil
}

// Get is ...
func Get(items []SourceIsins, dbpath string) error {

	// check input
	if err := checkListOfSourceIsins(items); err != nil {
		return err
	}

	// Workers
	ws := make([]*taskengine.Worker, 0, len(items))

	// WorkerTasks
	wts := make(taskengine.WorkerTasks)

	for _, item := range items {

		qg := quoteGetter[item.Source]

		// work function of the source
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
				r.Price = res.Price
				r.Currency = res.Currency
				if !res.Date.IsZero() {
					r.Date = &res.Date
				}
			}
			if err != nil {
				r.ErrMsg = err.Error()
				if e, ok := err.(*scrapers.Error); ok {
					r.Isin = e.Isin
					r.Source = e.Name
					r.URL = e.URL
				}
			}
			return r
		}

		// worker
		w := &taskengine.Worker{
			WorkerID:  taskengine.WorkerID(item.Source),
			Instances: item.Workers,
			Work:      wfn,
		}
		ws = append(ws, w)

		// Tasks
		ts := make(taskengine.Tasks, 0, len(item.Isins))
		for _, isin := range item.Isins {
			ts = append(ts, &taskGetQuote{
				isin: isin,
				url:  "",
				// proxy: item.Proxy,
			})
		}
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
