package scrapers

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter"
)

// Scraper interface
type Scraper interface {
	Source() string
	Client() *http.Client
	GetSearch(ctx context.Context, isin string) (*http.Request, error)
	ParseSearch(doc *goquery.Document, isin string) (string, error)
	GetInfo(ctx context.Context, isin, url string) (*http.Request, error)
	ParseInfo(doc *goquery.Document, isin string) (*ParseInfoResult, error)
}

// ParseInfoResult is ...
type ParseInfoResult struct {
	IsinStr     string
	PriceStr    string
	CurrencyStr string
	DateStr     string
	DateLayout  string
}

// quoteGetter is ...
type quoteGetter struct {
	Scraper
}

// NewQuoteGetter is ..
func NewQuoteGetter(scr Scraper) quotegetter.QuoteGetter {
	return &quoteGetter{scr}
}

// // Register register a new scrapers
// func Register(scr scrapers) {
// 	quotegetter.Register(newQuoteGetter(scr))
// }

// GetQuote is ...
func (qg *quoteGetter) GetQuote(ctx context.Context, isin, url string) (*quotegetter.Result, error) {
	return getQuote(ctx, isin, url, qg)
}

// getInfoFromDoc parse the info page and returns the result
func getInfoFromDoc(docInfo *goquery.Document, isin, url string, scr Scraper) (*quotegetter.Result, error) {
	var (
		pir *ParseInfoResult
		err error
	)

	// aux function
	theError := func(err error, typ ErrorType) (*quotegetter.Result, error) {
		e := &Error{
			ParseInfoResult: pir,
			source:          scr.Source(),
			isin:            isin,
			url:             url,
			err:             err,
			errType:         typ,
		}
		return nil, e
	}

	// parse the info document to get the results
	pir, err = scr.ParseInfo(docInfo, isin)
	if err != nil {
		errType := ParseInfoError
		if err == ErrNoResultFound {
			errType = NoResultFoundError
		}
		return theError(err, errType)
	}

	//check ISIN
	if isin != pir.IsinStr {
		return theError(ErrIsinMismatch, IsinMismatchError)
	}

	// parse price
	vPrice, err := parsePrice(pir.PriceStr)
	if err != nil {
		return theError(err, InvalidPriceError)
	}

	// parse date
	vDate, err := parseDate(pir.DateStr, pir.DateLayout)
	if err != nil {
		return theError(err, InvalidDateError)
	}

	r := &quotegetter.Result{
		Source:   scr.Source(),
		Isin:     isin,
		URL:      url,
		Price:    vPrice,
		Date:     vDate,
		Currency: quotegetter.NormalizeCurrency(pir.CurrencyStr),
	}
	return r, nil
}

func getQuote(ctx context.Context, isin, url string, scr Scraper) (*quotegetter.Result, error) {

	var (
		req  *http.Request
		resp *http.Response
		doc  *goquery.Document
		err  error
	)
	// aux function
	theError := func(err error, typ ErrorType) (*quotegetter.Result, error) {
		e := &Error{
			source:  scr.Source(),
			isin:    isin,
			url:     url,
			err:     err,
			errType: typ,
		}
		return nil, e
	}

	if scr == nil {
		return nil, fmt.Errorf("getQuote: scraper is nil")
	}

	if url == "" {
		// get the search page
		req, err = scr.GetSearch(ctx, isin)

		// reqSearch can be nil if the Info URL can be build from isin only
		if req != nil && err == nil {
			resp, err = quotegetter.DoHTTPRequest(scr.Client(), req)
		}
		if err != nil {
			return theError(err, GetSearchError)
		}

		// create goquery document only if respSearch != nil
		if resp != nil {
			defer resp.Body.Close()

			// set url to SearchURL for error reporting pourposes.
			// it will be overwritten in case of success finding InfoURL.
			url = resp.Request.URL.String()

			// docSearch, err = goquery.NewDocumentFromResponse(respSearch)
			doc, err = goquery.NewDocumentFromReader(resp.Body)
			// err != nil is handled below
		}

		if err == nil {
			// NOTE: docSearch can be nil
			//       if the url can be build from isin only
			url, err = scr.ParseSearch(doc, isin)

			if resp != nil && strings.HasPrefix(url, "/") {
				// prepend scheme://host from respSearch.Request.URL
				u, err := neturl.Parse(url)
				if err != nil {
					return theError(err, ParseSearchError)
				}

				url = resp.Request.URL.ResolveReference(u).String()
			}

		}

		if url == "" {
			err = ErrEmptyInfoURL
		}

		if err != nil {
			return theError(err, ParseSearchError)
		}
	}

	// check url (that is != "" )
	_, err = neturl.Parse(url)
	if err != nil {
		return theError(err, GetInfoError)
	}

	// get the info page
	req, err = scr.GetInfo(ctx, isin, url)
	if err == nil {
		if req == nil {
			return theError(ErrInfoRequestIsNil, GetInfoError)
		}
		resp, err = quotegetter.DoHTTPRequest(scr.Client(), req)
	}
	if err != nil {
		return theError(err, GetInfoError)
	}
	defer resp.Body.Close()

	// create goquery document
	// docInfo, err := goquery.NewDocumentFromResponse(respInfo)
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return theError(err, ParseInfoError)
	}

	return getInfoFromDoc(doc, isin, url, scr)

}

// ============================================================================
// aux functions

func parseDate(str, layout string) (time.Time, error) {
	var t time.Time
	if str == "" {
		return t, ErrDateNotFound
	}
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return t, err
	}
	return time.ParseInLocation(layout, str, loc)
}

func parsePrice(str string) (float32, error) {
	if str == "" {
		return 0.0, ErrPriceNotFound
	}
	price, err := strconv.ParseFloat(strings.Replace(str, ",", ".", 1), 32)
	return float32(price), err
}

// SplitPriceCurrency is ...
func SplitPriceCurrency(txt string, priceFirst bool) (priceStr string, currencyStr string, err error) {
	// split price and currency (11.49 EUR)

	// replace &nbsp; unicode char with space
	// NOTE: not needed with Split (needed with Split)
	// txt = strings.ReplaceAll(txt, "\u00a0", " ")

	// split the string
	a := strings.Fields(txt)
	if len(a) < 2 {
		priceStr = txt
		err = fmt.Errorf("Invalid price and currency string: %q", txt)
		return
	}

	var idxPrice int
	if !priceFirst {
		idxPrice++
	}

	priceStr = a[idxPrice]
	currencyStr = a[1-idxPrice]
	return
}
