package fondidocit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/quote/internal/quotegetter/scrapers"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/testingscraper"
)

func getTestScraper() scrapers.Scraper {
	return &scraper{"fondidocit", nil}
}

func checkError(t *testing.T, prefix string, found, expected error) bool {
	if found != expected {
		if expected == nil {
			t.Errorf("%s: unexpected error %q", prefix, found)
		} else if found == nil {
			t.Errorf("%s: expected error %q, found <nil>", prefix, expected)
		} else {
			t.Errorf("%s: expected error %q, expected %q", prefix, expected, found)
		}
		return true
	}
	return found != nil
}

func TestSource(t *testing.T) {
	const expected = "dummy"
	scr := &scraper{expected, nil}
	if actual := scr.Source(); actual != expected {
		t.Errorf("Source: expected %q, found %q", expected, actual)
	}
}

func TestGetSearch(t *testing.T) {
	scr := getTestScraper()
	testingscraper.TestGetSearch(t, "", scr)
}

func TestParseSearch(t *testing.T) {

	testCases := []struct {
		title string
		isin  string
		url   string
		err   error
		html  string
	}{
		{"ok",
			"IE00B4TG9K96",
			"/d/Ana/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg",
			nil,
			`<table>
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
</tr></table>`,
		},
		{"ko-isin",
			"IE00B4TG9XYZ",
			"",
			scrapers.ErrNoResultFound,
			`<table>
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
</tr></table>`,
		},
		{"ko-url",
			"IE00B4TG9K96",
			"",
			scrapers.ErrNoResultFound,
			`<table>
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
	<a fidacode="PIMDIEHI" purl="IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" href="">
		PIMCO Diversified Income E Dis EUR Hdg
	</a>
</td>
<td>
	IE00B4TG9K96
</td>
</tr></table>`,
		},

		{"ko-all",
			"IE00B4TG9ABC",
			"",
			scrapers.ErrNoResultFound,
			``,
		},
	}

	scr := getTestScraper()

	for _, tc := range testCases {
		prefix := fmt.Sprintf("ParseSearch[%s]", tc.title)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
		if checkError(t, prefix+": Goquery", err, nil) {
			continue
		}

		url, err := scr.ParseSearch(doc, tc.isin)
		if checkError(t, prefix, err, tc.err) {
			continue
		}

		if url != tc.url {
			t.Errorf("%s: URL expected %q, found %q", prefix, tc.url, url)
		}

	}

}

func TestGetInfo(t *testing.T) {
	scr := getTestScraper()
	testingscraper.TestGetInfo(t, "", scr)
}

/*

 */

func TestParseInfo(t *testing.T) {

	testCases := []struct {
		title    string
		priceStr string
		dateStr  string
		err      error
		html     string
	}{
		{
			title:    "ok",
			priceStr: "11,400",
			dateStr:  "22/09/2020",
			err:      nil,
			html: `<div class="page-header">
		<a href="/Confronto/Index/PIMDIEHI" style="float:right;margin-top:10px;" class="btn btn-default btn-sm btn-primary" ><i class="glyphicon glyphicon-plus"></i> Confronta</a>
		<h1>PIMCO Diversified Income E Dis EUR Hdg <small>IE00B4TG9K96</small></h1>
	</div>
	<div>
		<ul class="nav nav-pills nav-dett" id="mainTab">
			<li id="liDtAnag" >
				<a id="aDtAnag" href="/d/Index/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" >Anagrafica</a></li>
			<li id="liDtDocu" >
				<a id="aDtDocu" href="/d/Doc/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" >Documentazione</a></li>
			<li id="liDtAnls" class=active>
				<a id="aDtAnls" href="/d/Ana/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" >Analisi</a></li>
			<li id="liDtPort" >
				<a id="aDtPort" href="/d/Port/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" >Portafogli</a></li>
			<li id="liDtNews" class=hidden>
				<a id="aDtNews" href="/d/News/PIMDIEHI/IE00B4TG9K96_pimco-diversified-income-e-dis-eur-hdg" >Notizie</a></li>
		</ul>
	</div>
	<div class="dett-cont tab-content">
		<div class="row">
			<div class="col-md-5">
				<h4>Quotazioni</h4>
				<dl class="dl-horizontal">
					<dt>Frequenza di quotazione</dt>
						<dd>Giornaliero</dd>
					<dt>Valuta di quotazione</dt>
						<dd>Euro</dd>
					<dt>Ultimo aggiornamento</dt>
						<dd>22/09/2020</dd>
					<dt>Valore quota</dt>
						<dd>11,400</dd>
					<dt>Variazione (%)</dt>
						<dd><span class="value-neg">-0,18%</span></dd>
				</dl>
			</div>
		</div>
	</div>`,
		},

		{
			title:    "ko-price-date",
			priceStr: "",
			dateStr:  "",
			err:      scrapers.ErrNoResultFound,
			html: `<div class="page-header">
		<a href="/Confronto/Index/PIMDIEHI" style="float:right;margin-top:10px;" class="btn btn-default btn-sm btn-primary" ><i class="glyphicon glyphicon-plus"></i> Confronta</a>
		<h1>PIMCO Diversified Income E Dis EUR Hdg <small>IE00B4TG9K96</small></h1>
	</div>

	<div class="dett-cont tab-content">
		<div class="row">
			<div class="col-md-5">
				<h4>Quotazioni</h4>
				<dl class="dl-horizontal">
					<dt>Frequenza di quotazione</dt>
						<dd>Giornaliero</dd>
					<dt>Valuta di quotazione</dt>
						<dd>Euro</dd>
					<dt>Ultimo aggiornamento</dt>
						<dd></dd>
					<dt>Valore quota</dt>
						<dd></dd>
					<dt>Variazione (%)</dt>
						<dd><span class="value-neg">-0,18%</span></dd>
				</dl>
			</div>
		</div>
	</div>`,
		},
		{
			title:    "ko-price",
			priceStr: "",
			dateStr:  "22/09/2020",
			err:      nil,
			html: `<div class="page-header">
		<a href="/Confronto/Index/PIMDIEHI" style="float:right;margin-top:10px;" class="btn btn-default btn-sm btn-primary" ><i class="glyphicon glyphicon-plus"></i> Confronta</a>
		<h1>PIMCO Diversified Income E Dis EUR Hdg <small>IE00B4TG9K96</small></h1>
	</div>

	<div class="dett-cont tab-content">
		<div class="row">
			<div class="col-md-5">
				<h4>Quotazioni</h4>
				<dl class="dl-horizontal">
					<dt>Frequenza di quotazione</dt>
						<dd>Giornaliero</dd>
					<dt>Valuta di quotazione</dt>
						<dd>Euro</dd>
					<dt>Ultimo aggiornamento</dt>
						<dd>22/09/2020</dd>
					<dt>Valore quota</dt>
						<dd></dd>
					<dt>Variazione (%)</dt>
						<dd><span class="value-neg">-0,18%</span></dd>
				</dl>
			</div>
		</div>
	</div>`,
		},

		{
			title:    "ko-date",
			priceStr: "123",
			dateStr:  "",
			err:      nil,
			html: `<div class="page-header">
		<a href="/Confronto/Index/PIMDIEHI" style="float:right;margin-top:10px;" class="btn btn-default btn-sm btn-primary" ><i class="glyphicon glyphicon-plus"></i> Confronta</a>
		<h1>PIMCO Diversified Income E Dis EUR Hdg <small>IE00B4TG9K96</small></h1>
	</div>

	<div class="dett-cont tab-content">
		<div class="row">
			<div class="col-md-5">
				<h4>Quotazioni</h4>
				<dl class="dl-horizontal">
					<dt>Frequenza di quotazione</dt>
						<dd>Giornaliero</dd>
					<dt>Valuta di quotazione</dt>
						<dd>Euro</dd>
					<dt>Ultimo aggiornamento</dt>
						<dd></dd>
					<dt>Valore quota</dt>
						<dd>123</dd>
					<dt>Variazione (%)</dt>
						<dd><span class="value-neg">-0,18%</span></dd>
				</dl>
			</div>
		</div>
	</div>`,
		},
	}

	scr := getTestScraper()

	for _, tc := range testCases {
		prefix := fmt.Sprintf("ParseInfo[%s]", tc.title)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
		if checkError(t, prefix+": Goquery", err, nil) {
			continue
		}

		res, err := scr.ParseInfo(doc, "")
		if checkError(t, prefix, err, tc.err) {
			continue
		}

		if res.PriceStr != tc.priceStr {
			t.Errorf("%s: PriceStr: expected %q, found %q", prefix, tc.priceStr, res.PriceStr)
		}
		if res.DateStr != tc.dateStr {
			t.Errorf("%s: DateStr: expected %q, found %q", prefix, tc.dateStr, res.DateStr)
		}
	}
}
