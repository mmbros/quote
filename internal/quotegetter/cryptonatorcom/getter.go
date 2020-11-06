package cryptonatorcom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mmbros/quote/internal/quotegetter"
)

// getter gets cryptocurrrencies prices from cryptonator.com
type getter struct {
	name     string
	client   *http.Client
	currency string
}

type jsonTicker struct {
	Base   string
	Target string
	Price  string
	Volume string
	Change string
}
type jsonResult struct {
	Ticker    jsonTicker
	Timestamp int64
	Success   bool
	Error     string
}

// NewQuoteGetter creates a new QuoteGetter
// that gets stock/fund prices from cryptonator.com
func NewQuoteGetter(name string, client *http.Client, currency string) quotegetter.QuoteGetter {
	return &getter{name, client, currency}
}

// Name returns the name of the scraper
func (g *getter) Source() string {
	return g.name
}

// Name returns the name of the scraper
func (g *getter) Client() *http.Client {
	return g.client
}

// GetQuote ....
func (g *getter) GetQuote(ctx context.Context, crypto, url string) (*quotegetter.Result, error) {

	if url == "" {
		url = fmt.Sprintf("https://api.cryptonator.com/api/ticker/%s-%s",
			strings.ToLower(crypto),
			strings.ToLower(g.currency))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	res, err := quotegetter.DoHTTPRequest(g.client, req)
	if err != nil {
		return nil, err
	}

	// body, err := ioutil.ReadAll(res.Body)
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	r, err := g.parseJSON(body)
	if err != nil {
		e := &quotegetter.Error{
			Isin:   crypto,
			URL:    url,
			Source: g.Source(),
			Err:    err,
		}
		return nil, e
	}
	r.URL = url

	return r, nil
}

func (g *getter) parseJSON(body []byte) (*quotegetter.Result, error) {

	var res jsonResult

	err := json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Success {
		price64, err := strconv.ParseFloat(res.Ticker.Price, 32)
		if err != nil {
			return nil, err
		}

		r := &quotegetter.Result{
			Isin:     res.Ticker.Base,
			Currency: res.Ticker.Target,
			Source:   g.Source(),
			Date:     time.Unix(res.Timestamp, 0),
			Price:    float32(price64),
		}
		return r, nil
	}

	return nil, errors.New(res.Error)
}
