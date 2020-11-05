package quotegetter

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/mmbros/quote/internal/quotetesting"
)

func TestDoHTTPRequestWithTimeout(t *testing.T) {
	const (
		timeout = 50
		delay   = 100
	)
	server := quotetesting.NewTestServer()
	defer server.Close()

	// context
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
	defer cancel()

	// url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %q", err)
	}
	q := u.Query()
	q.Set("delay", strconv.Itoa(delay))
	q.Set("isin", "ISIN00001234")
	q.Set("date", "2020-09-26")
	q.Set("price", "100.01")
	q.Set("currency", "EUR")
	u.RawQuery = q.Encode()

	// request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %q", err)
	}

	// do
	_, err = DoHTTPRequest(nil, req)

	// check
	if err == nil {
		t.Error("Expected error, got success")
		return
	}
	expected := context.DeadlineExceeded
	if uErr, ok := err.(*url.Error); !ok || uErr.Err != expected {
		t.Errorf("Expected error %q, got %q", expected, err)
	}
}

func TestDoHTTPRequestWithCancel(t *testing.T) {
	const (
		timeout = 50
		delay   = 100
	)
	server := quotetesting.NewTestServer()
	defer server.Close()

	// context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(timeout * time.Millisecond)
		cancel()
	}()

	// url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %q", err)
	}
	q := u.Query()
	q.Set("delay", strconv.Itoa(delay))
	q.Set("isin", "ISIN00001234")
	q.Set("date", "2020-09-26")
	q.Set("price", "100.01")
	q.Set("currency", "EUR")
	u.RawQuery = q.Encode()

	// request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %q", err)
	}

	// do
	_, err = DoHTTPRequest(nil, req)

	// check
	if err == nil {
		t.Error("Expected error, got success")
		return
	}
	expected := context.Canceled
	if uErr, ok := err.(*url.Error); !ok || uErr.Err != expected {
		t.Errorf("Expected error %q, got %q", expected, err)
	}
}

func TestDoHTTPRequestOK(t *testing.T) {

	server := quotetesting.NewTestServer()
	defer server.Close()

	// context
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %q", err)
	}
	q := u.Query()
	q.Set("delay", "100")
	q.Set("isin", "ISIN00001234")
	q.Set("date", "2020-09-26")
	q.Set("price", "100.01")
	q.Set("currency", "EUR")
	u.RawQuery = q.Encode()

	// request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %q", err)
	}

	// do
	resp, err := DoHTTPRequest(nil, req)

	// check
	if err != nil {
		t.Fatalf("doHTTPRequest: %q", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Read Body: %q", err)
	}

	t.Log(string(body))
}

func TestDoHTTPRequestKO(t *testing.T) {
	server := quotetesting.NewTestServer()
	defer server.Close()

	// url
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %q", err)
	}
	q := u.Query()
	q.Set("isin", "ISIN00001234")
	q.Set("code", strconv.Itoa(http.StatusInternalServerError))
	u.RawQuery = q.Encode()

	// request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		t.Fatalf("NewRequest: %q", err)
	}

	// do
	_, err = DoHTTPRequest(nil, req)

	// check
	if err == nil {
		t.Error("Expected error, got success")
		return
	}
	t.Log(err)
	// t.Fail()
}
