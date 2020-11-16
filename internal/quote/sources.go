package quote

import (
	"net/http"
	"sort"

	"github.com/mmbros/quote/internal/quotegetter"
	"github.com/mmbros/quote/internal/quotegetter/jsons/cryptonatorcom"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/fondidocit"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/fundsquarenet"
	"github.com/mmbros/quote/internal/quotegetter/scrapers/morningstarit"
)

type fnNewQuoteGetter func(string, *http.Client) quotegetter.QuoteGetter

var availableSources map[string]fnNewQuoteGetter

func init() {

	fnCryptonatorcom := func(currency string) fnNewQuoteGetter {
		return func(name string, client *http.Client) quotegetter.QuoteGetter {
			return cryptonatorcom.NewQuoteGetter(name, client, currency)
		}
	}

	availableSources = map[string]fnNewQuoteGetter{
		"fondidocit":         fondidocit.NewQuoteGetter,
		"morningstarit":      morningstarit.NewQuoteGetter,
		"fundsquarenet":      fundsquarenet.NewQuoteGetter,
		"cryptonatorcom-EUR": fnCryptonatorcom("EUR"),
		// "cryptonatorcom-USD": fnCryptonatorcom("USD"),
	}

}

func initQuoteGetters(src []*SourceIsins) (map[string]quotegetter.QuoteGetter, error) {
	quoteGetter := make(map[string]quotegetter.QuoteGetter)

	proxyClient := map[string]*http.Client{}

	for _, s := range src {
		name := s.Source

		client, ok := proxyClient[s.Proxy]
		if !ok {
			client, err := quotegetter.DefaultClient(s.Proxy)
			if err != nil {
				return nil, err
			}
			proxyClient[s.Proxy] = client
		}

		fn := availableSources[name]
		if fn == nil {
			panic("invalid source: " + name)
		}
		quoteGetter[name] = fn(name, client)
	}

	return quoteGetter, nil
}

// getSources returns a list of the names of the available quoteGetters.
func getSources() []string {

	list := make([]string, 0, len(availableSources))
	for name := range availableSources {
		list = append(list, name)
	}

	return list
}

// Sources returns a sorted list of the names of the avaliable quoteGetters.
func Sources() []string {
	list := getSources()
	sort.Strings(list)
	return list
}
