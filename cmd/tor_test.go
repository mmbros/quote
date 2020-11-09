package cmd

import (
	"testing"
)

func TestGetProxy(t *testing.T) {
	torProxy := "socks5://127.0.0.1:9050"
	cases := []struct {
		title  string
		arg    string
		passed bool
		cfg    string
		env    string
		want   string
	}{
		{
			title:  "arg",
			arg:    torProxy,
			passed: true,
			cfg:    `proxy: skipped`,
			want:   torProxy,
		},
		{
			title:  "arg empty",
			arg:    "",
			passed: true,
			cfg:    `proxy: skipped`,
			want:   "",
		},
		{
			title:  "arg-ref",
			arg:    "tor",
			passed: true,
			cfg: `
proxies:
- proxy: tor
  url: ` + torProxy,
			want: torProxy,
		},
		{
			title: "cfg",
			cfg:   `proxy: ` + torProxy,
			want:  torProxy,
		},
		{
			title: "cfg-ref",
			cfg: `
proxy: tor
proxies:
- proxy: tor
  url: ` + torProxy,
			want: torProxy,
		},
	}

	for _, c := range cases {
		initViperConfig(c.cfg)

		checkTorCmd.ResetFlags()
		checkTorCmd.Flags().StringP(nameProxy, "p", "", "proxy to test the TOR connection")
		if c.passed {
			checkTorCmd.ParseFlags([]string{"--proxy", c.arg})
		}

		p := getProxy(checkTorCmd)
		if p != c.want {
			t.Errorf("%s: want %q, got %q", c.title, c.want, p)
		}
	}

}
