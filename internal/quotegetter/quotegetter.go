package quotegetter

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// QuoteGetter interface
type QuoteGetter interface {
	Source() string
	Client() *http.Client
	GetQuote(ctx context.Context, isin, url string) (*Result, error)
}

// Result represents the info returned by the GetQuote function
type Result struct {
	Source   string
	Isin     string
	URL      string
	Price    float32
	Currency string
	Date     time.Time
}

// Error is
type Error struct {
	Source string
	Isin   string
	URL    string
	Err    error
}

// NormalizeCurrency return the standard ISO4217 representation
// of the known currency
func NormalizeCurrency(currency string) string {
	if strings.EqualFold(currency, "euro") {
		return "EUR"
	}
	return currency
}

func (e *Error) Error() string {
	return e.Err.Error()
}
