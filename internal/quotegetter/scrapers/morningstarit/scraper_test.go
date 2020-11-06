package morningstarit

import (
	"testing"

	"github.com/mmbros/quote/internal/quotegetter/scrapers"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/testingscraper"
)

func getTestScraper() scrapers.Scraper {
	return &scraper{"morningstarit", nil}
}

func TestSource(t *testing.T) {
	const name = "dummy"
	scr := &scraper{name, nil}
	if nameFound := scr.Source(); nameFound != name {
		t.Errorf("Source: found %q, expected %q", nameFound, name)
	}
}

func TestGetSearch(t *testing.T) {
	scr := getTestScraper()
	testingscraper.TestGetSearch(t, "", scr)
}

func TestParseSearch(t *testing.T) {
	testCases := []struct {
		isin     string
		filename string
		url      string
		err      error
	}{
		{"IE00B4TG9K96", "morningstar.it/search|IE00B4TG9K96|ok.html", "/it/funds/snapshot/snapshot.aspx?id=F000005GUM", nil},
		{"IE00B4TG9KAA", "morningstar.it/search|IE00B4TG9KAA|ko.html", "", scrapers.ErrNoResultFound},
	}

	scr := getTestScraper()

	for _, tc := range testCases {
		doc, err := testingscraper.GetDoc(tc.filename)
		if err != nil {
			t.Error(tc.filename, err)
			continue
		}

		url, err := scr.ParseSearch(doc, tc.isin)
		if url != tc.url {
			t.Errorf("[%s] ParseSearch: URL found %q, expected %q", scr.Source(), url, tc.url)
		}
		if err != tc.err {
			t.Errorf("[%s] ParseSearch: ERR found %q, expected %q", scr.Source(), err, tc.err)
		}
	}
}

func TestGetInfo(t *testing.T) {
	scr := getTestScraper()
	testingscraper.TestGetInfo(t, "", scr)
}

func TestParseInfo(t *testing.T) {

	testCases := []struct {
		filename string
		priceStr string
		dateStr  string
		isinStr  string
	}{
		{"morningstar.it/info|IT0005247157|ok.html", "126,370", "28/08/2020", "IT0005247157"},
	}

	scr := getTestScraper()

	for _, tc := range testCases {

		doc, err := testingscraper.GetDoc(tc.filename)
		if err != nil {
			t.Error(tc.filename, err)
			continue
		}

		res, err := scr.ParseInfo(doc, "")
		if err != nil {
			if tc.priceStr != "" {
				t.Errorf("[%s] Unexpected error %q", tc.filename, err)
			}
			if err != scrapers.ErrNoResultFound {
				t.Errorf("[%s] Unexpected error %q, expected %q", tc.filename, err, scrapers.ErrNoResultFound)
			}

			continue
		}

		t.Logf("[%s] -> %q", tc.filename, res)

		if res.PriceStr != tc.priceStr {
			t.Errorf("[%s] PriceStr: expected %q, found %q", tc.filename, tc.priceStr, res.PriceStr)
		}
		if res.DateStr != tc.dateStr {
			t.Errorf("[%s] DateStr: expected %q, found %q", tc.filename, tc.dateStr, res.DateStr)
		}
		if res.IsinStr != tc.isinStr {
			t.Errorf("[%s] IsinStr: expected %q, found %q", tc.filename, tc.isinStr, res.IsinStr)
		}
	}
}
