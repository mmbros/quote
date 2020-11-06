package fondidocit

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/internal/quotegetter/scrapers"
)

// scraper gets stock/fund prices from fondidoc.it
type scraper struct {
	name   string
	client *http.Client
}

// NewQuoteGetter creates a new QuoteGetter
// that gets stock/fund prices from fondidoc.it
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

// GetSearch creates the http.Request to get the search page for the specified `isin`.
// It returns the http.Response or nil if the scraper can build the url of the info page
// directly from the `isin`.
// The response document will be parsed by ParseSearch to extract the info url.
func (s *scraper) GetSearch(ctx context.Context, isin string) (*http.Request, error) {
	url := fmt.Sprintf("https://www.fondidoc.it/Ricerca/Res?txt=%s&tipi=&societa=&pag=0&sort=&sortDir=&fldis=&nview=20&viewMode=anls&filters=&pir=0'", isin)
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

// ParseSearch parse the html of the search page to find the URL of the info page.
// `doc` can be nil if the url of the info page can be build directly from the `isin`.
// It returns the url of the info page.
func (s *scraper) ParseSearch(doc *goquery.Document, isin string) (string, error) {
	/*
		   <tr>
		       <td>
		           <div style="position:relative;">
		               <button class="btn btn-default btn-xs" data-toggle="dropdown"><i class="glyphicon glyphicon-plus"></i></button>
		               <ul class="dropdown-menu">
		                   <li><a href="/Confronto/Index/PIMDIEHI">Aggiungi a confronto</a></li>
		               </ul>
		           </div>
		       </td>
		       <td>
		           <a fidacode="PIMDIEHI" purl="IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" href="/d/Ana/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg">
		               PIMCO Diversified Income E Dis EUR Hdg
		           </a>
		       </td>
		       <td>
		           IE00B4TG9K96
			   </td>
			</tr>
	*/

	var url string
	var found bool

	doc.Find("tr").EachWithBreak(func(iTR int, sTR *goquery.Selection) bool {

		sTR.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
			switch i {
			case 1:
				url = s.Find("a").AttrOr("href", "")
			case 2:
				theIsin := strings.TrimSpace(s.Text())
				found = (theIsin == isin) && (url != "")
				return false
			}
			return true
		})
		return !found
	})

	if !found {
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
		<div class="page-header">
		    <a href="/Confronto/Index/PIMDIEHI" style="float:right;margin-top:10px;" class="btn btn-default btn-sm btn-primary" ><i class="glyphicon glyphicon-plus"></i> Confronta</a>
		    <h1>PIMCO Diversified Income E Dis EUR Hdg <small>IE00B4TG9K96</small></h1>
		</div>

		div.dett-cont dd
		[0] Giornaliero
		[1] Euro
		[2] 22/09/2020
		[3] 11,400
		[4] -0,18%
	*/

	r := new(scrapers.ParseInfoResult)
	r.DateLayout = "02/01/2006"

	r.IsinStr = doc.Find("div.page-header small").Text()

	doc.Find("div.dett-cont dd").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch i {
		case 1:
			r.CurrencyStr = s.Text()
		case 2:
			r.DateStr = s.Text()
		case 3:
			r.PriceStr = s.Text()
			return false
		}
		return true
	})

	if r.DateStr == "" && r.PriceStr == "" {
		return r, scrapers.ErrNoResultFound
	}
	return r, nil
}
