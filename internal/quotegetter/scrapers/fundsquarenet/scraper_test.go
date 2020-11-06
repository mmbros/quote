package fundsquarenet

import (
	"strings"
	"testing"

	"github.com/mmbros/quote/internal/quotegetter/scrapers"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/testingscraper"
)

func getTestScraper() scrapers.Scraper {
	return &scraper{"fundsquarenet", nil}
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
	const isin = "<ISIN>"
	scr := getTestScraper()
	url, err := scr.ParseSearch(nil, isin)
	if err != nil {
		t.Errorf("ParseSearch: %v", err)
	}
	if !strings.Contains(url, isin) {
		t.Errorf("ParseSearch: returned url %q does not contain isin %q", url, isin)
	}
}

func TestGetInfo(t *testing.T) {
	scr := getTestScraper()
	req, _ := testingscraper.TestGetInfo(t, "", scr)
	if req != nil {
		referer := req.Header.Get("Referer")
		isin := testingscraper.TestIsin
		if !strings.Contains(referer, isin) {
			t.Errorf("GetInfo: referer %q not contains isin %q", referer, isin)
		}
	}
}

func TestParseInfo(t *testing.T) {

	testCases := []struct {
		filename string
		priceStr string
		dateStr  string
	}{
		{"fundsquare.net/info|IE00B4TG9K96|ok.html", "11.49", "11/09/2020"},
		{"fundsquare.net/info|IE00B4TG9K96|unavailable.html", "11.47", "18/09/2020"},
		{"fundsquare.net/info|IT0005247157|not-found.html", "", ""},
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
		t.Log(tc.filename, "->", res)

		if res.PriceStr != tc.priceStr {
			t.Errorf("[%s] PriceStr: expected %q, found %q", tc.filename, tc.priceStr, res.PriceStr)
		}
		if res.DateStr != tc.dateStr {
			t.Errorf("[%s] DateStr: expected %q, found %q", tc.filename, tc.dateStr, res.DateStr)
		}
	}
}
