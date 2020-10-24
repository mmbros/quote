package htmlquotescraper

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is the http.Client used for the quote requests.
var Client *http.Client

func init() {
	Client = DefaultClient("")
}

// DefaultClient xxx
func DefaultClient(proxy string) *http.Client {
	// tr := &http.Transport{}
	tr := http.DefaultTransport.(*http.Transport).Clone()

	if len(proxy) > 0 {
		// Parse proxy URL string to a URL type
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			panic(fmt.Sprintf("Error parsing proxy URL: %q. %v", proxy, err))
		}
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}

// doHTTPRequest executes the http request.
func doHTTPRequest(req *http.Request) (*http.Response, error) {
	resp, err := Client.Do(req)
	if (err == nil) && (resp.StatusCode != http.StatusOK) {
		err = fmt.Errorf("Get %q with response status = %v", req.URL, resp.Status)
	}
	return resp, err
}
