package quotegetter

import (
	"context"
	"fmt"
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

// Error is the interface that must be matched by all quotegetter errors
type Error interface {
	Source() string
	Isin() string
	URL() string
	Error() string
	Unwrap() error
}

// getterError is a base implementation of quotegetter.Error interface
type getterError struct {
	source string
	isin   string
	url    string
	err    error
}

// NewError creates a new *jsonGetterError
func NewError(source, isin, url string, err error) error {
	return &getterError{source, isin, url, err}
}

// Source returns the Source of the error
func (e *getterError) Source() string { return e.source }

// Isin returns the Isin of the error
func (e *getterError) Isin() string { return e.isin }

// URL returns the URL of the error
func (e *getterError) URL() string { return e.url }

// Unwrap returns the inner error
func (e *getterError) Unwrap() error { return e.err }

// Error return the string representation of the error
func (e *getterError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("Unknown error getting quote of isin %q from source %q", e.isin, e.source)
	}
	return e.err.Error()
}

// NormalizeCurrency return the standard ISO4217 representation
// of the known currency
func NormalizeCurrency(currency string) string {
	if strings.EqualFold(currency, "euro") {
		return "EUR"
	}
	return currency
}
