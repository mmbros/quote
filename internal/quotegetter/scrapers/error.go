package scrapers

import (
	"errors"
	"fmt"

	"github.com/mmbros/quote/internal/quotegetter"
)

// ErrorType is ...
//go:generate stringer -type=ErrorType
type ErrorType int

//  ErrorType enum
const (
	Success ErrorType = iota
	NoResultFoundError
	IsinMismatchError
	GetSearchError
	ParseSearchError
	GetInfoError
	ParseInfoError
	PriceNotFoundError
	InvalidPriceError
	DateNotFoundError
	InvalidDateError
	IsinNotFoundError
)

// Error  is ...
// FIXME use quetegetter.Error
// type Error struct {
// 	*ParseInfoResult
// 	Type ErrorType
// 	Name string
// 	Isin string
// 	URL  string
// 	Err  error
// }

type quotegetterError quotegetter.Error

// Error is
type Error struct {
	*ParseInfoResult
	errType ErrorType
	source  string
	isin    string
	url     string
	err     error
}

// Source returns the Source of the error
func (e *Error) Source() string { return e.source }

// Isin returns the Isin of the error
func (e *Error) Isin() string { return e.isin }

// URL returns the URL of the error
func (e *Error) URL() string { return e.url }

// Unwrap returns the inner error
func (e *Error) Unwrap() error { return e.err }

// Error return the string representation of the error
func (e *Error) Error() string {

	var sInnerErr string
	if e.err != nil {
		sInnerErr = e.err.Error()
	}

	switch e.errType {
	case IsinMismatchError:
		return fmt.Sprintf("%s: expected %q, found %q", sInnerErr, e.isin, e.IsinStr)
	case NoResultFoundError, InvalidPriceError:
		return fmt.Sprintf("%s for isin %q", sInnerErr, e.isin)
	default:
		return fmt.Sprintf("%s: %s", e.errType.String(), sInnerErr)
	}

}

// Errors
var (
	ErrNoResultFound    = errors.New("no result found")
	ErrIsinMismatch     = errors.New("isin mismatch")
	ErrEmptyInfoURL     = errors.New("parse search returned an empty info URL")
	ErrInfoRequestIsNil = errors.New("info request is nil")
	ErrPriceNotFound    = errors.New("price not found")
	ErrDateNotFound     = errors.New("date not found")
)
