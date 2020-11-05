package scrapers

import (
	"errors"
	"fmt"
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
type Error struct {
	*ParseInfoResult
	Type ErrorType
	Name string
	Isin string
	URL  string
	Err  error
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {

	var sInnerErr string
	if e.Err != nil {
		sInnerErr = e.Err.Error()
	}

	switch e.Type {
	case IsinMismatchError:
		return fmt.Sprintf("%s: expected %q, found %q in url %q", sInnerErr, e.Isin, e.IsinStr, e.URL)
	case NoResultFoundError, InvalidPriceError:
		return fmt.Sprintf("%s for isin %q in url %q", sInnerErr, e.Isin, e.URL)
	default:
		return fmt.Sprintf("%s: %s", e.Type.String(), sInnerErr)
	}

}

// Errors
var (
	ErrNoResultFound    = errors.New("no result found")
	ErrIsinMismatch     = errors.New("isin mismatch")
	ErrEmptyInfoURL     = errors.New("parse search returned an empty info URL")
	ErrInfoRequestIsNil = errors.New("info request is nil")
	ErrPriceNotFound    = errors.New("Price not found")
	ErrDateNotFound     = errors.New("Date not found")
)
