package quotegetter

import (
	"context"
	"time"
)

// QuoteGetter interface
type QuoteGetter interface {
	Name() string
	GetQuote(ctx context.Context, isin, url string) (*Result, error)
}

// Result represents the info returned by the GetQuote function
type Result struct {
	Name     string
	Isin     string
	URL      string
	Price    float32
	Currency string
	Date     time.Time
}
