package quote

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter"
)

// TorCheck checks if a Tor connection is used,
// retrieving the "https://check.torproject.org" page.
// It returns:
//  - bool:   true if Tor is used, false otherwise
//  - string: the message contained in the html page
//  - error:  if the message cannot be determined
func TorCheck(proxy string) (bool, string, error) {
	// URL to fetch
	var webURL string = "https://check.torproject.org"

	client, err := quotegetter.DefaultClient(proxy)
	if err != nil {
		return false, "", err
	}
	// Make request
	resp, err := client.Get(webURL)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, "", err
	}

	msg := strings.TrimSpace(doc.Find("h1").Text())
	if msg == "" {
		return false, "", fmt.Errorf("can't determine if you are using Tor")
	}

	return msg == "Congratulations. This browser is configured to use Tor.", msg, nil
}
