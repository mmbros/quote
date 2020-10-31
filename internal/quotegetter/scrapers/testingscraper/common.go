package testingscraper

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// exported constants
const (
	TestIsin    = "ISIN00001234"
	TestInfoURL = "http://127.0.0.1/info/ISIN00001234"
)

type scraperTest interface {
	// Name() string
	GetSearch(ctx context.Context, isin string) (*http.Request, error)
	// ParseSearch(doc *goquery.Document, isin string) (string, error)
	GetInfo(ctx context.Context, isin, url string) (*http.Request, error)
	// ParseInfo(doc *goquery.Document) (*ParseInfoResult, error)
}

// CheckError is ...
func CheckError(t *testing.T, prefix string, found, expected error) bool {
	if found != expected {
		if expected == nil {
			t.Errorf("%s: unexpected error %q", prefix, found)
		} else if found == nil {
			t.Errorf("%s: expected error %q, found <nil>", prefix, expected)
		} else {
			// check string and inner error
			for inner := found; inner != nil; {

				// check by string
				if inner.Error() == expected.Error() {
					return true
				}
				// check inner error (if found)
				inner = errors.Unwrap(inner)
				if inner == expected {
					return true
				}
			}

			t.Errorf("%s: expected error %q, found %q", prefix, expected, found)
		}
		return true
	}
	return found != nil
}

// TestGetSearch is
func TestGetSearch(t *testing.T, prefix string, scr scraperTest) (*http.Request, error) {
	req, err := scr.GetSearch(context.Background(), TestIsin)
	if err != nil {
		t.Errorf("%s: %v", prefix, err)
	}
	return req, err
}

// TestGetInfo is
func TestGetInfo(t *testing.T, prefix string, scr scraperTest) (*http.Request, error) {

	if prefix == "" {
		prefix = "GetInfo"
	}
	url := TestInfoURL
	req, err := scr.GetInfo(context.Background(), TestIsin, url)

	if err != nil {
		t.Errorf("%s: %v", prefix, err)
	} else if req == nil {
		t.Errorf("%s: req is %v", prefix, nil)
	} else if requrl := req.URL.String(); requrl != url {
		t.Errorf("%s: invalid URL: expected %q, found %q", prefix, url, requrl)
	}
	return req, err
}

// NewDocumentFromFile returns the goquery.Document created by a local html file
func NewDocumentFromFile(path string) (*goquery.Document, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return goquery.NewDocumentFromReader(f)
}

// NewDocumentFromZipFile returns the goquery.Document created by a zipped html file
func NewDocumentFromZipFile(file *zip.File) (*goquery.Document, error) {
	fc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fc.Close()

	return goquery.NewDocumentFromReader(fc)
}

func getBaseDir() string {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	base := strings.SplitAfter(currentWorkingDirectory, "quote")[0]
	return base
}

// getFullPath returns the full path to the file "test/internal/quotescaper/<relpath>".
func getFullPath(relpath string) string {
	base := getBaseDir()
	return filepath.Join(base, "/test/internal/quotescraper", relpath)
}

// GetDoc returns the goquery.Document created by
// the local html file "test/internal/quotescaper/<relpath>".
func GetDoc(relpath string) (*goquery.Document, error) {
	// use file system first
	fullpath := getFullPath(relpath)
	doc, err := NewDocumentFromFile(fullpath)
	if err == nil {
		return doc, nil
	}

	// then use zip file
	return zipGetDoc(relpath)
}

func zipGetDoc(relpath string) (*goquery.Document, error) {
	// retrieve zip file full path
	zipfile := filepath.Join(getBaseDir(), "/test/internal/quotescraper.zip")

	// open zip file
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	fname := filepath.Join("quotescraper", relpath)

	// Iterate through the files in the archive,
	for _, f := range r.File {
		if f.Name == fname {
			return NewDocumentFromZipFile(f)
		}
	}

	return nil, fmt.Errorf("File %q not found in %q", relpath, zipfile)
}
