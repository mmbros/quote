package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mmbros/quote/internal/quote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {

	dataYaml := []byte(`
# quote configuration file
database: /home/user/quote.sqlite3
workers: 2
proxy: proxy1
proxies:
    proxy1: socks5://localhost:9051
    none: ""
isins:
    isin1:
        sources: [source1]
sources:
    source1:
        proxy: none
        disabled: y
`)

	dataToml := []byte(`
# quote configuration file

database = "/home/user/quote.sqlite3"

workers = 2

proxy = "proxy1"

[proxies]
proxy1 = "socks5://localhost:9051"
none = ""

[isins]

[isins.isin1]
sources = ["source1"]
disabled = false

[sources]

[sources.source1]
proxy = "none"
disabled = true
`)

	dataJSON := []byte(`{
	"database": "/home/user/quote.sqlite3",
	"workers": 2,
	"proxy": "proxy1",
	"proxies": {
	  "none": "",
	  "proxy1": "socks5://localhost:9051"
	},
	"sources": {
	  "source1": {
		 "disabled": true,
		 "proxy": "none"
	  }
	},
	"isins": {
	  "isin1": {
		  "sources": [
			"source1"
		  ]
	  }
	}
  }
  `)

	expected := &Config{
		Database: "/home/user/quote.sqlite3",
		Workers:  2,
		Proxy:    "proxy1",
		Proxies: map[string]string{
			"proxy1": "socks5://localhost:9051",
			"none":   "",
		},
		Isins: map[string]*isinItem{
			"isin1": {
				Sources: []string{"source1"},
			},
		},
		Sources: map[string]*sourceItem{
			"source1": {
				Proxy:    "none",
				Disabled: true,
			},
		},
	}

	cases := []struct {
		fmt  string
		data []byte
	}{
		{"yaml", dataYaml},
		{"yml", dataYaml},
		{"", dataYaml},
		{"toml", dataToml},
		{"", dataToml},
		{"json", dataJSON},
		{"", dataJSON},
	}

	var cfg *Config

	for _, c := range cases {
		cfg = &Config{}
		err := unmarshal(c.data, cfg, c.fmt)
		msg := fmt.Sprintf("case with fmt %q, len(data)=%d", c.fmt, len(c.data))
		if assert.NoError(t, err, msg) {
			assert.Equal(t, expected, cfg, msg)
		}
	}
}

func TestUnmarshalError(t *testing.T) {

	cases := []struct {
		fmt     string
		strdata string
		errmsg  string
	}{
		{"", "x y z", "Unknown format"},
		{"config", "x y z", "Unsupported format"},
	}

	var cfg *Config

	for _, c := range cases {
		cfg = &Config{}
		err := unmarshal([]byte(c.strdata), cfg, c.fmt)
		msg := fmt.Sprintf("case with fmt %q", c.fmt)
		if assert.Error(t, err, msg) {
			assert.Contains(t, err.Error(), c.errmsg, msg)
		}
	}
}

func TestGetFileFormat(t *testing.T) {

	cases := []struct {
		path string
		fmt  string
		want string
	}{
		{"/home/user/config.YAML", "", "yaml"},
		{"config.toml", "EXT", "ext"},
		{"config", "", ""},
	}

	for _, c := range cases {
		got := getFileFormat(c.path, c.fmt)
		assert.Equal(t, c.want, got, c)
	}
}

func initAppGetArgs(options string) (*appArgs, error) {
	args := &appArgs{}
	cmd := initCommandGet(args)
	fs := cmd.FlagSet(nil)
	err := fs.Parse(strings.Split(options, " "))

	return args, err
}

func TestWorkers(t *testing.T) {
	availableSources := []string{"source1"}

	cases := map[string]struct {
		argtxt string
		cfgtxt string
		want   []*quote.SourceIsins
		errmsg string
	}{
		"none": {
			want: []*quote.SourceIsins{},
		},
		"workers = 0": {
			argtxt: "-w 0 -i isin1",
			errmsg: "workers must be greater than zero",
		},
		"workers < 0": {
			argtxt: "-w -10 -i isin1",
			errmsg: "workers must be greater than zero",
		},
		"workers > 0": {
			argtxt: "-w 10 -i isin1",
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: 10, Isins: []string{"isin1"}},
			},
		},
		"default with args": {
			argtxt: "-i isin1",
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: defaultWorkers, Isins: []string{"isin1"}},
			},
		},
		"default with cfg": {
			cfgtxt: `isins:
  isin1:
    sources: [source1]`,
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: defaultWorkers, Isins: []string{"isin1"}},
			},
		},
		"args with source1:0": {
			argtxt: "-i isin1 -s source1:0",
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: defaultWorkers, Isins: []string{"isin1"}},
			},
		},
		"args with source1:-1": {
			argtxt: "-i isin1 -s source1:-1",
			errmsg: "workers must be greater than zero",
		},
		"args with source1:100": {
			argtxt: "-i isin1 -s source1:100",
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: 100, Isins: []string{"isin1"}},
			},
		},
		"args with source1#100": {
			argtxt: "-i isin1 -s source1#100",
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: 100, Isins: []string{"isin1"}},
			},
		},
		"args with source1/100": {
			argtxt: "-i isin1 -s source1/100",
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: 100, Isins: []string{"isin1"}},
			},
		},
		"cfg source with workers=0": {
			argtxt: "--config-type=yaml",
			cfgtxt: `
isins:
  isin1:
sources:
  source1:
    workers: 0
`,
			want: []*quote.SourceIsins{
				{Source: "source1", Workers: defaultWorkers, Isins: []string{"isin1"}},
			},
		},
		"cfg source with workers=-1": {
			argtxt: "--config-type=yaml",
			cfgtxt: `
isins:
  isin1:
sources:
  source1:
    workers: -1
`,
			errmsg: "workers must be greater than zero (source \"source1\" has workers=-1)",
		},
		"cfg with workers=-1": {
			argtxt: "--config-type=yaml",
			cfgtxt: `
workers: -10
isins:
  isin1:
`,
			errmsg: "workers must be greater than zero (workers=-10)",
		},
	}
	for title, c := range cases {

		cfg := &Config{}
		args, _ := initAppGetArgs(c.argtxt)
		err := cfg.auxGetConfig([]byte(c.cfgtxt), args, availableSources)

		if c.errmsg != "" {
			if assert.Error(t, err, title) {
				assert.Contains(t, err.Error(), c.errmsg, title)
			}
		} else {
			if assert.NoError(t, err, title) {
				got := cfg.SourceIsinsList()
				assert.ElementsMatch(t, c.want, got, title)
			}
		}
	}
}

func TestProxy(t *testing.T) {

	availableSources := []string{"source1", "source2", "source3"}

	yaml1 := `
proxy: common

isins:
  isin1:

proxies:
  none: ""
  common: http://common
  proxy2: http://proxy2

sources:
  source1:
    proxy: http://proxy1
  source2:
    proxy: none
`

	cases := map[string]struct {
		argtxt string
		cfgtxt string
		want1  string
		want2  string
		want3  string
		errmsg string
	}{
		"args only": {
			argtxt: "-i isin1",
		},
		"args only with proxy": {
			argtxt: "-i isin1 -p x://y",
			want1:  "x://y",
			want2:  "x://y",
			want3:  "x://y",
		},
		"args invalid proxy": {
			argtxt: "-i isin1 -p x://\\",
			errmsg: "invalid proxy",
		},
		"args ignored unused invalid proxy": {
			argtxt: "-p x://\\",
		},
		"cfg only": {
			cfgtxt: yaml1,
			want1:  "http://proxy1",
			want2:  "",
			want3:  "http://common",
		},
		"cfg with arg proxy": {
			argtxt: "-p http://args",
			cfgtxt: yaml1,
			want1:  "http://proxy1",
			want2:  "",
			want3:  "http://args",
		},
		"cfg with arg proxy-ref": {
			argtxt: "-p proxy2",
			cfgtxt: yaml1,
			want1:  "http://proxy1",
			want2:  "",
			want3:  "http://proxy2",
		},
		"cfg with arg proxy-ref to none": {
			argtxt: "-p none",
			cfgtxt: yaml1,
			want1:  "http://proxy1",
			want2:  "",
			want3:  "",
		},
	}
	for title, c := range cases {

		cfg := &Config{}
		args, err := initAppGetArgs(c.argtxt)
		require.NoError(t, err)
		err = cfg.auxGetConfig([]byte(c.cfgtxt), args, availableSources)

		if c.errmsg != "" {
			if assert.Error(t, err, title) {
				assert.Contains(t, err.Error(), c.errmsg, title)
			}
		} else {
			if assert.NoError(t, err, title) {
				got := cfg.SourceIsinsList()
				mgot := map[string]string{}
				for _, si := range got {
					mgot[si.Source] = si.Proxy
				}
				mwant := map[string]string{
					"source1": c.want1,
					"source2": c.want2,
					"source3": c.want3,
				}

				for s, want := range mwant {
					if want != mgot[s] {
						t.Errorf("case %q: %q: want %q, got %q", title, s, want, mgot[s])
					}
				}
			}
		}
	}
}
