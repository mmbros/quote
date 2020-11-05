package scrapers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/testingscraper"
	"github.com/mmbros/quote/internal/quotetesting"
)

type testCaseGetQuote struct {
	title    string
	price    float32
	currency string
	date     time.Time
	err      error
}

var testCasesGetQuote = map[string]*testCaseGetQuote{
	"ISIN00000001": {
		title:    "ok",
		price:    12.34,
		currency: "EUR",
		date:     time.Date(2020, time.February, 23, 0, 0, 0, 0, time.UTC),
	},
	"ISIN00000002": {
		title:    "ok, abs-url",
		price:    12.34,
		currency: "EUR",
		date:     time.Date(2020, time.February, 23, 0, 0, 0, 0, time.UTC),
	},
	"ISIN00000003": {
		title:    "ko-no-price",
		price:    0,
		currency: "EUR",
		date:     time.Date(2020, time.February, 23, 0, 0, 0, 0, time.UTC),
		err:      ErrPriceNotFound,
	},
	"ISIN00000004": {
		title:    "ko-no-date",
		price:    123,
		currency: "EUR",
		err:      ErrDateNotFound,
	},
	"ISIN00000005": {
		title: "ko, no-info-result",
		err:   ErrNoResultFound,
	},
	"ISIN00000006": {
		title:    "ko, isin-mismatch",
		price:    12.34,
		currency: "EUR",
		date:     time.Date(2020, time.February, 23, 0, 0, 0, 0, time.UTC),
		err:      ErrIsinMismatch,
	},
	"ISIN00000007": {
		title: "ko, no-info-url",
		err:   ErrEmptyInfoURL,
	},
	"ISIN00000008": {
		title: "ko, info-invalid-url",
		err:   errors.New("ParseSearchError: parse \"/\\nnewline\": net/url: invalid control character in URL"),
	},
	"ISIN00000009": {
		title: "ko, abs-url, info-invalid-url",
		err:   errors.New("net/url: invalid control character in URL"),
	},
	"ISIN00000010": {
		title: "ko-get-info-nil",
		err:   ErrInfoRequestIsNil,
	},
	"ISIN00000011": {
		title: "ko-get-info-500",
		err:   errors.New("GetInfoError: Response status = 500 Internal Server Error"),
	},
	"ISIN00000012": {
		title: "ko-parse-info",
		err:   ErrNoResultFound,
	},
	"ISIN00000013": {
		title: "ko-timeout-get-search",
		err:   context.DeadlineExceeded,
	},
	"ISIN00000014": {
		title: "ko-timeout-get-info",
		err:   context.DeadlineExceeded,
	},
}

type testScraper struct {
	name      string
	serverURL string
}

func newTestScraper(name, url string) *testScraper {
	return &testScraper{name, url}
}

func (scr testScraper) Name() string {
	return scr.name
}

func (scr testScraper) Client() *http.Client {
	return nil
}
func (scr testScraper) GetSearch(ctx context.Context, isin string) (*http.Request, error) {

	tc := testCasesGetQuote[isin]
	if tc == nil {
		panic("testScraper.GetSearch: isin not found: " + isin)
	}

	u, err := url.Parse(scr.serverURL)
	if err != nil {
		panic("testScraper.GetSearch: " + err.Error())
	}
	q := u.Query()
	q.Set("op", "search")
	q.Set("isin", isin)

	if strings.Contains(tc.title, "ko-timeout-get-search") {
		q.Set("delay", "500")
	}
	u.RawQuery = q.Encode()

	return http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

}

func (scr testScraper) ParseSearch(doc *goquery.Document, isin string) (string, error) {
	var err error
	tc := testCasesGetQuote[isin]
	if tc == nil {
		panic("testScraper.ParseSearch: isin not found: " + isin)
	}

	if strings.Contains(tc.title, "no-info-url") {
		return "", nil
	}
	if strings.Contains(tc.title, "info-invalid-url") {
		s := `/
newline`
		if strings.Contains(tc.title, "abs-url") {
			s = scr.serverURL + s
		}
		return s, nil
	}

	rootURL := "/"
	if strings.Contains(tc.title, "abs-url") {
		rootURL = scr.serverURL
	}

	u, err := url.Parse(rootURL)
	if err != nil {
		panic("testScraper.ParseSearch: " + err.Error())
	}
	q := u.Query()
	q.Set("op", "info")

	if strings.Contains(tc.title, "isin-mismatch") {
		q.Set("isin", "ISINMISMATCH")
	} else {
		q.Set("isin", isin)
	}

	if !tc.date.IsZero() {
		q.Set("date", tc.date.Format(time.RFC3339)[:10])
	}
	if tc.price > 0.000001 {
		q.Set("price", fmt.Sprintf("%f", tc.price))
	}
	if len(tc.currency) > 0 {
		q.Set("currency", tc.currency)
	}
	if strings.Contains(tc.title, "ko-get-info-500") {
		q.Set("code", "500")
	}
	if strings.Contains(tc.title, "ko-timeout-get-info") {
		q.Set("delay", "500")
	}
	u.RawQuery = q.Encode()

	infoURL := u.String()

	if strings.Contains(tc.title, "no-info-result") {
		err = ErrNoResultFound
	}

	return infoURL, err
}

func (scr testScraper) GetInfo(ctx context.Context, isin, url string) (*http.Request, error) {

	tc := testCasesGetQuote[isin]
	if strings.Contains(tc.title, "ko-get-info-nil") {
		return nil, nil
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (scr testScraper) ParseInfo(doc *goquery.Document, isin string) (*ParseInfoResult, error) {

	r := new(ParseInfoResult)
	r.DateLayout = "2006-01-02"

	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		th := s.Find("th").Text()
		td := s.Find("td").Text()
		switch th {
		case "isin":
			r.IsinStr = td
		case "date":
			r.DateStr = td
		case "currency":
			r.CurrencyStr = td
		case "price":
			r.PriceStr = td
		}
	})

	if tc, ok := testCasesGetQuote[r.IsinStr]; ok {

		if strings.Contains(tc.title, "ko-parse-info") {
			return nil, ErrNoResultFound
		}
	}

	return r, nil
}

func TestGetQuote(t *testing.T) {

	server := quotetesting.NewTestServer()
	defer server.Close()

	// ctx := context.Background()

	// scraper
	scr := newTestScraper("localhost", server.URL)

	for isin, tc := range testCasesGetQuote {
		// if isin != "ISIN00000011" {
		// 	continue
		// }

		// context
		ctx := context.Background()
		if strings.Contains(tc.title, "timeout") {
			newctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()
			ctx = newctx
		}

		res, err := getQuote(ctx, isin, "", scr)

		prefix := fmt.Sprintf("GetQuote[%s]", tc.title)
		if testingscraper.CheckError(t, prefix, err, tc.err) {
			continue
		}

		if res.Currency != tc.currency {
			t.Errorf("%s: Currency: expected %s, found %s", prefix, tc.currency, res.Currency)
		}
		if res.Price != tc.price {
			t.Errorf("%s: Price: expected %f, found %f", prefix, tc.price, res.Price)
		}

	}

}

func TestSplitPriceCurrency(t *testing.T) {
	testCases := []struct {
		txt         string
		priceFirst  bool
		priceStr    string
		currencyStr string
		err         error
	}{
		{"12.34 EUR XXX YY", true, "12.34", "EUR", nil},
		{"12.34\u00a0EUR XXX YY", true, "12.34", "EUR", nil},
		{"  USD \u00a0  12.34  XXX YY Z", false, "12.34", "USD", nil},
		{"String", true, "", "", errors.New("Invalid price and currency string: \"String\"")},
		{"", false, "", "", errors.New("Invalid price and currency string: \"\"")},
	}

	prefix := "ParsePriceCurrency"
	for _, tc := range testCases {

		priceStr, currencyStr, err := SplitPriceCurrency(tc.txt, tc.priceFirst)
		if testingscraper.CheckError(t, prefix, err, tc.err) {
			continue
		}
		if currencyStr != tc.currencyStr {
			t.Errorf("%s: currencyStr: expected %s, found %s", prefix, tc.currencyStr, currencyStr)
		}
		if priceStr != tc.priceStr {
			t.Errorf("%s: priceStr: expected %s, found %s", prefix, tc.priceStr, priceStr)
		}
	}
}

func TestNewQuoteGetter(t *testing.T) {

	scr := newTestScraper("localhost", "http://127.0.0.1")

	qg := NewQuoteGetter(scr)
	if qg == nil {
		t.Errorf("NewQuoteGetter: returned nil")
	}
}

func TestQuoteGetterGetQuote(t *testing.T) {
	server := quotetesting.NewTestServer()
	defer server.Close()

	scr := newTestScraper("localhost", server.URL)

	qg := NewQuoteGetter(scr)
	if qg == nil {
		t.Errorf("NewQuoteGetter: returned nil")
		return
	}

	for isin, tc := range testCasesGetQuote {

		// context
		ctx := context.Background()
		if strings.Contains(tc.title, "timeout") {
			newctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()
			ctx = newctx
		}

		res, err := qg.GetQuote(ctx, isin, "")

		prefix := fmt.Sprintf("GetQuote[%s]", tc.title)
		if testingscraper.CheckError(t, prefix, err, tc.err) {
			continue
		}

		if res.Currency != tc.currency {
			t.Errorf("%s: Currency: expected %s, found %s", prefix, tc.currency, res.Currency)
		}
		if res.Price != tc.price {
			t.Errorf("%s: Price: expected %f, found %f", prefix, tc.price, res.Price)
		}

	}

}
