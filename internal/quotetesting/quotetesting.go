package quotetesting

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"
)

// NewTestServer create a new httptest server that returns a response
// build on the request parameters.
// The response in an html page with a table that prints
// kew/value pairs of query parameters.
// Special parameters:
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
<h1>Test Server Result</h1>
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
