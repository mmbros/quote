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
	Name     string    `json:"source"`
	Isin     string    `json:"isin"`
	URL      string    `json:"url"`
	Price    float32   `json:"price"`
	Currency string    `json:"currency"`
	Date     time.Time `json:"date"`
}
