package htmlquotescraper

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// doHTTPRequest executes the http request.
// It cancels the request in case the `ctx` is cancelled.
func doHTTPRequest(req *http.Request) (*http.Response, error) {

	// Specify proxy ip and port
	var theProxy string = "socks5://127.0.0.1:9050"

	// make the request
	// tr := &http.Transport{}
	tr := http.DefaultTransport.(*http.Transport).Clone()

	if len(theProxy) > 0 {
		// Parse proxy URL string to a URL type
		proxyURL, err := url.Parse(theProxy)
		if err != nil {
			panic(fmt.Sprintf("Error parsing proxy URL: %q. %v", theProxy, err))
		}
		tr.Proxy = http.ProxyURL(proxyURL)

	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if (err == nil) && (resp.StatusCode != http.StatusOK) {
		err = fmt.Errorf("Response status = %v", resp.Status)
	}

	return resp, err
}

/*
// doHTTPRequest executes the http request.
// It cancels the request in case the `ctx` is cancelled.
func doHTTPRequestOLD(ctx context.Context, req *http.Request) (*http.Response, error) {

	type result struct {
		resp *http.Response
		err  error
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// make the request
		tr := &http.Transport{}
		client := &http.Client{
			Transport: tr,
			Timeout:   10 * time.Second,
		}

		c := make(chan result, 1)

		go func() {
			resp, err := client.Do(req)
			c <- result{resp: resp, err: err}
		}()

		// wait for the result or the cancel signal
		select {
		case <-ctx.Done():
			tr.CancelRequest(req)
			<-c // Wait for client.Do
			return nil, ctx.Err()
		case r := <-c:

			if (r.err == nil) && (r.resp.StatusCode != http.StatusOK) {
				r.err = fmt.Errorf("Response status = %v", r.resp.Status)
			}

			return r.resp, r.err
		}
	}
}
*/
