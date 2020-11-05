package cryptonatorcom

import (
	"context"
	"encoding/json"
	"testing"
)

// func TestGetJson(t *testing.T) {

// 	res, err := http.Get("https://api.cryptonator.com/api/ticker/btc-eur")
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// 	body, err := ioutil.ReadAll(res.Body)
// 	res.Body.Close()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// 	t.Logf(string(body))
// 	// t.Fail()

// 	// {"ticker":{"base":"BTC","target":"EUR","price":"11872.29709977","volume":"9489.21251997","change":"56.52524067"},"timestamp":1604159942,"success":true,"error":""}
// }

func TestParseJson(t *testing.T) {
	type Ticker struct {
		Base   string
		Target string
		Price  string
		Volume string
		Change string
	}
	type Result struct {
		Ticker    Ticker
		Timestamp int64
		Success   bool
		Error     string
	}

	eq := func(title, expected, got string) {
		if expected != got {
			t.Errorf("%s: expected %q, got %q", title, expected, got)
		}
	}
	eqi := func(title string, expected, got int64) {
		if expected != got {
			t.Errorf("%s: expected %d, got %d", title, expected, got)
		}
	}

	var res Result
	body := `{"ticker":{"base":"BTC","target":"EUR","price":"11872.29709977","volume":"9489.21251997","change":"56.52524067"},"timestamp":1604159942,"success":true,"error":""}`

	err := json.Unmarshal([]byte(body), &res)
	if err != nil {
		t.Fatal(err)
	}

	eq("base", "BTC", res.Ticker.Base)
	eq("target", "EUR", res.Ticker.Target)
	eq("price", "11872.29709977", res.Ticker.Price)
	eqi("timestamp", 1604159942, res.Timestamp)

}

func TestGetQuote(t *testing.T) {
	g := NewQuoteGetter("cryptonator-eur", nil, "EUR")

	ctx := context.Background()
	r, err := g.GetQuote(ctx, "BTC", "")

	if err != nil {
		t.Fatalf(err.Error())
	}

	t.Logf("OK %v", r)

	// BTC2 -> Pair not found
	// EURO -> Pair not found
}
