package htmlquotescraper

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

// NewTestServer create a new httptest server that returns a response build on the request parameters.
// special parameters:
//   delay: number of msec to sleep before returning the response
//   code: returned http status
func NewTestServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()

		// code
		code, _ := strconv.Atoi(query.Get("code"))
		if code == 0 {
			code = http.StatusOK
		}

		// delay
		delaymsec, _ := strconv.Atoi(query.Get("delay"))
		if delaymsec > 0 {
			time.Sleep(time.Duration(delaymsec) * time.Millisecond)
		}

		if code != http.StatusOK {
			// set the status code
			http.Error(w, http.StatusText(code), code)
			return
		}

		fmt.Fprint(w, `<html>
<head>
<title>Test Server Result</title>
</head>
<body>
<table>
`)
		for k, v := range query {
			fmt.Fprintf(w, "<tr><th>%s</th><td>%s</td></tr>\n", k, v[0])
		}
		fmt.Fprint(w, `</table>
</body>
</html>`)

	}))

	return server
}

func TestDoHTTPRequestWithTimeout(t *testing.T) {
	const (
		timeout = 50
		delay   = 100
	)
	server := NewTestServer()
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
	_, err = doHTTPRequest(req)

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
	server := NewTestServer()
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
	_, err = doHTTPRequest(req)

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

	server := NewTestServer()
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
	resp, err := doHTTPRequest(req)

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
	server := NewTestServer()
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
	_, err = doHTTPRequest(req)

	// check
	if err == nil {
		t.Error("Expected error, got success")
		return
	}
	t.Log(err)
	// t.Fail()
}
