package quotegetterdb

import (
	"fmt"
	"testing"
	"time"
)

const (
	isin1 = "ISIN00001111"
	isin2 = "ISIN99992222"

	source1 = "quotesource1.com"
	source2 = "quotesource2.it"
)

func testURL(source, isin string) string {
	return fmt.Sprintf("http://127.0.0.1/%s/%s/", source, isin)
}

var loc = time.Local

var records = []*QuoteRecord{
	{
		isin:      isin1,
		source:    source1,
		price:     10.1,
		currency:  "USD",
		date:      time.Date(2020, 01, 01, 0, 0, 0, 0, loc),
		timestamp: time.Date(2020, 01, 01, 10, 11, 0, 0, loc),
		url:       testURL(source1, isin1),
	},
	{
		isin:      isin1,
		source:    source1,
		timestamp: time.Date(2020, 01, 02, 10, 22, 0, 0, loc),
		errmsg:    "Isin not found",
	},
	{
		isin:      isin1,
		source:    source1,
		price:     10.3,
		currency:  "USD",
		date:      time.Date(2020, 01, 03, 0, 0, 0, 0, loc),
		timestamp: time.Date(2020, 01, 03, 10, 33, 0, 0, loc),
		url:       testURL(source1, isin1),
	},
	{
		isin:      isin1,
		source:    source2,
		price:     10.22,
		currency:  "EUR",
		date:      time.Date(2020, 02, 01, 0, 0, 0, 0, loc),
		timestamp: time.Date(2020, 02, 01, 0, 0, 0, 0, loc),
		url:       testURL(source2, isin1),
	},
	{
		isin:      isin1,
		source:    source2,
		timestamp: time.Date(2020, 02, 02, 0, 0, 0, 0, loc),
		errmsg:    "Isin not found",
	},
	{
		isin:      isin1,
		source:    source2,
		timestamp: time.Date(2020, 02, 03, 0, 0, 0, 0, loc),
		errmsg:    "Isin not found",
	},
	{
		isin:      isin1,
		source:    source2,
		timestamp: time.Date(2020, 02, 04, 0, 0, 0, 0, loc),
		errmsg:    "Isin not found",
	},
	{
		isin:      isin2,
		source:    source1,
		timestamp: time.Date(2020, 02, 03, 0, 0, 0, 0, loc),
		errmsg:    "Isin not found",
	}}

// const dbpath = ":memory:"
const dbpath = "/tmp/quote.sqlite3"

func mustOpenDB() *QuoteDatabase {

	qdb, err := Open(dbpath)
	if err != nil {
		panic(err)
	}
	return qdb
}

func TestOpen(t *testing.T) {

	// create the database
	qdb, err := Open(dbpath)
	if err != nil {
		t.Fatal(err)
	}
	qdb.Close()
}

func TestInsertQuotes(t *testing.T) {
	qdb := mustOpenDB()
	defer qdb.Close()

	err := qdb.InsertQuotes(records...)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSelectLastQuotes(t *testing.T) {
	qdb := mustOpenDB()
	defer qdb.Close()

	err := qdb.InsertQuotes(records...)
	if err != nil {
		t.Fatal(err)
	}

	res, err := qdb.SelectLastQuotes()
	if err != nil {
		t.Fatal(err)
	}
	for j, r := range res {

		t.Logf("[%d] %v\n", j, r)
	}
	t.Fail()
}

/*
func TestExtractPath(t *testing.T) {
	testCases := []struct {
		dns      string
		expected string
	}{
		{"/tmp/finanze/db.sqlite3", "/tmp/finanze"},
		{"file:/tmp/finanze/db.sqlite3", "/tmp/finanze"},
		{"file:/tmp/finanze/db.sqlite3?cache=shared", "/tmp/finanze"},
		{":memory:", ""},
		{"file::memory:", ""},
		{"file::memory:?cache=shared", ""},
		{"file:/tmp/finanze/db.sqlite3?cache=shared&mode=memory", ""},
		{"file:/tmp/finanze/db.sqlite3?mode=memory&cache=shared", ""},
	}

	for _, tc := range testCases {
		res := extractDir(tc.dns)
		if res != tc.expected {
			t.Errorf("extractPath(%q): got %q, expected %q", tc.dns, res, tc.expected)
		}
	}
}
*/
