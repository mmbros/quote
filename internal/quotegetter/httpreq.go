package quotegetter

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DefaultClient xxx
func DefaultClient(proxy string) (*http.Client, error) {
	// tr := &http.Transport{}
	tr := http.DefaultTransport.(*http.Transport).Clone()

	if len(proxy) > 0 {
		// Parse proxy URL string to a URL type
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
			// panic(fmt.Sprintf("Error parsing proxy URL: %q. %v", proxy, err))
		}
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
	return client, nil
}

// DoHTTPRequest executes the http request.
func DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	if client == nil {
		client, _ = DefaultClient("")
	}
	resp, err := client.Do(req)
	if (err == nil) && (resp.StatusCode != http.StatusOK) {
		err = fmt.Errorf("%s response status = %v", req.Method, resp.Status)
	}
	return resp, err
}
