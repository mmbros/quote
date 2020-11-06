package morningstarit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/internal/quotegetter/scrapers"
)

// scraper gets stock/fund prices from www.morningstar.it
type scraper struct {
	name   string
	client *http.Client
}

// NewQuoteGetter creates a new QuoteGetter
// that gets stock/fund prices from www.morningstar.it
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
	url := fmt.Sprintf("https://www.morningstar.it/it/funds/SecuritySearchResults.aspx?search=%s&type=", isin)
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

// ParseSearch parse the html of the search page to find the URL of the info page.
// `doc` can be nil if the url of the info page can be build directly from the `isin`.
// It returns the url of the info page.
func (s *scraper) ParseSearch(doc *goquery.Document, isin string) (string, error) {
	/*
		### SUCCESS
		<table id="ctl00_MainContent_fundTable" style="border-collapse:collapse;" cellspacing="0" cellpadding="0" border="0">
		  <tbody>
		    <tr class="searchGridHeader">
			  <th>Nome</th>
			  <th>ISIN</th>
		    </tr>
		    <tr class="gridItem">
			  <td class="msDataText searchLink">
			    <a href="/it/funds/snapshot/snapshot.aspx?id=F000005GUM">PIMCO GIS Divers Inc E EURH Inc</a>
			  </td>
			  <td class="msDataText searchIsin">
				<span>IE00B4TG9K96</span>
			  </td>
		    </tr>
		  </tbody>
		</table>

		### NO RESULT FOUND
		<span id="ctl00_ctl00_MainContent_Layout_1MainContent_lblEmptyDataMessage">Nessun risultato trovato.</span>
	*/
	url, ok := doc.Find("#ctl00_MainContent_fundTable td.searchLink > a").Attr("href")
	if !ok {
		return "", scrapers.ErrNoResultFound
	}

	return url, nil
}

// GetInfo executes the http GET of the `url` of info page for the specified `isin`.
// `url` and `isin` must be defined.
// The response document will be parsed by ParseInfo to extract the info url.
func (s *scraper) GetInfo(ctx context.Context, isin, url string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

// ParseInfo is ...
func (s *scraper) ParseInfo(doc *goquery.Document, isin string) (*scrapers.ParseInfoResult, error) {
	/*
		<table class="snapshotTextColor snapshotTextFontStyle snapshotTable overviewKeyStatsTable" border="0">
		  <tbody>
			<tr>
			  <td class="titleBarHeading" colspan="3">Sintesi</td></tr>
			<tr>
			  <td class="line heading">NAV<span class="heading"><br>28/08/2020</span></td>
			  <td class="line">&nbsp;</td>
			  <td class="line text">EUR&nbsp;126,370</td></tr>
			<tr>
			  <td class="line heading">Var.Ultima Quotazione</td>
			  <td class="line">&nbsp;</td>
			  <td class="line text">0,24%</td></tr>
			<tr>
			  <td class="line heading">Categoria Morningstar™</td>
			  <td class="line">&nbsp;</td>
			  <td class="line value text"><a href="https://www.morningstar.it/it/fundquickrank/default.aspx?category=EUCA000640" style="width:100%!important;">Azionari Italia</a></td></tr>
			<tr>
			  <td class="line heading">Categoria Assogestioni</td>
			  <td class="line">&nbsp;</td>
			  <td class="line text">Azionari Italia</td></tr>
			<tr>
			  <td class="line heading">Isin</td>
			  <td class="line">&nbsp;</td>
			  <td class="line text">IT0005247157</td></tr>
			<tr><td class="line heading">Fund Size (Mil)<span class="heading"><br>26/06/2020</span></td><td class="line">&nbsp;</td><td class="line text">EUR&nbsp;17,32</td></tr><tr><td class="line heading">Share Class Size (Mil)<span class="heading"><br>26/06/2020</span></td><td class="line">&nbsp;</td><td class="line text">EUR&nbsp;5,06</td></tr><tr><td class="line heading">Entrata (max)</td><td class="line">&nbsp;</td><td class="line text">-</td></tr><tr><td class="line heading"><a href="http://www.morningstar.it/it/glossary/121049/spese-correnti-ongoing-charge.aspx" style="text-decoration:underline;">Spese
	*/

	r := new(scrapers.ParseInfoResult)
	r.DateLayout = "02/01/2006"
	var txtPriceCurrency string

	doc.Find("table.overviewKeyStatsTable td").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch i {
		case 1:
			r.DateStr = s.Find("span").Text()
		case 3:
			txtPriceCurrency = s.Text()
		case 15:
			r.IsinStr = s.Text()
			return false
		}
		return true
	})

	// split price and currency (EUR 126,370)
	var errPrice error
	r.PriceStr, r.CurrencyStr, errPrice = scrapers.SplitPriceCurrency(txtPriceCurrency, false)

	return r, errPrice
}
