package fundsquarenet

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/internal/quotegetter/scrapers"
)

// scraper gets stock/fund prices from fundsquare.net
type scraper struct {
	name   string
	client *http.Client
}

// NewQuoteGetter creates a new QuoteGetter
// that gets stock/fund prices from fundsquare.net
func NewQuoteGetter(name string, client *http.Client) quotegetter.QuoteGetter {
	return scrapers.NewQuoteGetter(&scraper{name, client})
}

// Name returns the name of the scraper
func (s *scraper) Source() string {
	return s.name
}

// Client returns the http.Client of the scraper
func (s *scraper) Client() *http.Client {
	return s.client
}

// GetSearch executes the http GET of the search page for the specified `isin`.
// It returns the http.Response or nil if the scraper can build the url of the info page
// directly from the `isin`.
// The response document will be parsed by ParseSearch to extract the info url.
func (s *scraper) GetSearch(ctx context.Context, isin string) (*http.Request, error) {
	return nil, nil
}

// ParseSearch parse the html of the search page to find the URL of the info page.
// `doc` can be nil if the url of the info page can be build directly from the `isin`.
// It returns the url of the info page.
func (s *scraper) ParseSearch(doc *goquery.Document, isin string) (string, error) {
	url := fmt.Sprintf("https://www.fundsquare.net/search-results?ajaxContentView=renderContent"+
		"&=undefined&search=%s&isISIN=O&lang=EN&fastSearch=O", isin)
	return url, nil
}

// GetInfo is ...
func (s *scraper) GetInfo(ctx context.Context, isin, url string) (*http.Request, error) {

	// headers of the http request
	headers := map[string]string{
		"User-Agent":       "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0",
		"Accept":           "text/html;type=ajax",
		"Accept-Language":  "en-US,en;q=0.5",
		"X-Requested-With": "XMLHttpRequest",
		"DNT":              "1",
		"Connection":       "keep-alive",
		"Referer":          "https://www.fundsquare.net/search-results?fastSearch=O&isISIN=O&search=" + isin,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// set the request's headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// ParseInfo is ...
func (s *scraper) ParseInfo(doc *goquery.Document, isin string) (*scrapers.ParseInfoResult, error) {
	// SCENARIO OK
	// -----------
	// <div id="content">
	//  <table style="width: 100%">
	//   <tr>
	//     <td>
	//      <span style="font-weight: bold;">IE00B4TG9K96</span>
	//      &nbsp;&nbsp;PIMCO GIS Diversified Income Fund E Hgd EUR Dis&nbsp;&nbsp;</td>
	//     <td></td></tr></table>
	//  <table width="85%">
	//   <tr>
	//    <td width="30%">Last NAV</td>
	//    <td width="15%">11/09/2020</td>
	//    <td width="55%">
	//     <span class="surligneorange">11.49 EUR</span>&nbsp;

	// SCENARIO Last NAV status = Unavailable
	// --------------------------------------
	// <div id="content">
	// 	<table style="width: 100%">
	// 		<tbody>
	// 			<tr>
	// 				<td><span style="font-weight: bold;">IE00B4TG9K96</span>  PIMCO GIS Diversified Income Fund E Hgd
	// 					EUR Dis  </td>
	// 				<td></td>
	// 			</tr>
	// 		</tbody>
	// 	</table>
	// 	<table width="85%">
	// 		<tbody>
	// 			<tr>
	// 				<td width="30%">Last NAV status</td>
	// 				<td width="70%">Unavailable - Closed Market / Bank Holiday  (from 21/09/2020  to 21/09/2020)</td>
	// 			</tr>
	// 		</tbody>
	// 	</table>
	// 	<table width="85%">
	// 		<tbody>
	// 			<tr>
	// 				<td width="30%">Previous NAV</td>
	// 				<td width="15%">18/09/2020</td>
	// 				<td width="55%"><span class="surligneorange">11.47 EUR</span> <span
	// 						style="color:#DD0000;text-align:left;padding:4px 0;"> -0.17  % <img
	// 							src="/images/share/variationNegative.gif" style="vertical-align:middle;" /></span></td>
	// 			</tr>
	// 		</tbody>
	// 	</table>

	// SCENARIO No result
	// ------------------
	// <div class="box-message-info" style="">
	// <h3 class="table01-title">
	// 	<big>Search result</big>
	// </h3>
	// <table width="100%">
	// 	<tbody>
	// 		<tr>
	// 			<td valign="middle" align="center">
	// 				<img src="/images/share/research.gif" alt="info"/>
	// 			</td>
	// 			<td valign="middle" align="center">
	// 				<span class="title">
	// 					<div class="contenu"><p class="zero"></p><p><span class="surligneorange"> No result</span> produced by your request.</p><p>Please<span class="surligneorange"> modify your search criteria</span>.</p><input type="submit" id="valida" name="valider" onclick="window.location=&#39;/search?fastSearch=O&amp;isISIN=O&amp;search=IT0005247157&#39;" class="back_search_w btn_w" value="Back to search"/><br class="clear_r"/><p></p></div><p class="ps">Number of results : <span>0</span></p>
	// 				</span>

	r := new(scrapers.ParseInfoResult)
	r.DateLayout = "02/01/2006"
	var txtPriceCurrency string

	isLastNavAvailable := false

	doc.Find("div#content table td").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch i {
		case 0:
			r.IsinStr = s.Find("span").Text()
		case 3:
			r.DateStr = s.Text()
			isLastNavAvailable = !strings.HasPrefix(r.DateStr, "Unavailable")
		case 4:
			if isLastNavAvailable {
				txtPriceCurrency = s.Find("span").Text() // 11.49 EUR
				return false
			}
		case 5:
			// Previous NAV
			r.DateStr = s.Text()
		case 6:
			// Previous NAV
			txtPriceCurrency = s.Find("span").Text() // 11.49 EUR
			return false
		}
		return true
	})

	if txtPriceCurrency == "" {
		// check for "No result produced by your request.""
		s := strings.TrimSpace(doc.Find("div.contenu span.surligneorange").Text())
		if strings.HasPrefix(s, "No result") {
			return r, scrapers.ErrNoResultFound
		}
	}

	// split price and currency (11.49 EUR)
	var errPrice error
	r.PriceStr, r.CurrencyStr, errPrice = scrapers.SplitPriceCurrency(txtPriceCurrency, true)

	return r, errPrice
}
