package simpleflag

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type cmdOptions struct {
	config  String
	proxy   String
	workers Int
	isins   Strings
	sources Strings
	dryrun  Bool
}

func (opts *cmdOptions) Clear() {
	var cleared cmdOptions
	*opts = cleared

	// default values
	// opts.workers.value = 1
}

func TestApp(t *testing.T) {

	opts := &cmdOptions{}

	app := App{
		Name:  "quote",
		Usage: "Usage: quote [commands]",
		Commands: []*Command{
			{
				Names: "get,g",
				Usage: `Usage: quote get [options]

Options:
  -c, --config path       config file (default is $HOME/.quote.yaml)
  -i, --isins strings     list of isins to get the quotes
  -n, --dry-run           perform a trial run with no request/updates made
  -p, --proxy url         default proxy
  -s, --sources strings   list of sources to get the quotes from
  -w, --workers int       number of workers (default 1)
`,
				Flags: []*Flag{
					{&opts.isins, "i,isin"},
					{&opts.sources, "s,sources"},
					{&opts.config, "c,config"},
					{&opts.proxy, "p,proxy"},
					{&opts.workers, "w,workers"},
					{&opts.dryrun, "n,dryrun,dry-run"},
				},
			},
			{
				Names: "tor",
				Usage: `Usage: quote tor [options]

Options:
  -c, --config path    config file (default is $HOME/.quote.yaml)
  -p, --proxy url      proxy to test the Tor network
`,
				Flags: []*Flag{
					{&opts.config, "c,config"},
					{&opts.proxy, "p,proxy"},
				},
			},
		},
	}

	cases := []struct {
		title    string
		args     []string
		expected *cmdOptions
		errmsg   string
	}{
		{
			title:  "no arguments",
			args:   []string{},
			errmsg: "no arguments",
		},
		{
			title:  "app help",
			args:   []string{"-h"},
			errmsg: "help requested",
		},
		{
			title:  "no command",
			args:   []string{"--dummy"},
			errmsg: "flag provided but not defined",
		},
		{
			title: "get no workers",
			args:  []string{"get", "-s", "source1", "--proxy", "url"},
			expected: &cmdOptions{
				sources: Strings{"source1"},
				proxy:   String{"url", true},
			},
		},
		{
			title: "get dry-run",
			args:  []string{"get", "-i", "isin1,isin2", "--isin", "isin3", "--w=5", "--dry-run=1"},
			expected: &cmdOptions{
				workers: Int{5, true},
				dryrun:  Bool{true, true},
				isins:   Strings{"isin1", "isin2", "isin3"},
			},
		},
		{
			title:  "get help",
			args:   []string{"get", "--help", "--w=5"},
			errmsg: "help requested",
		},
		{
			title:  "tor help",
			args:   []string{"tor", "-help"},
			errmsg: "help requested",
		},
		{
			title:  "command not found",
			args:   []string{"dummy", "-h"},
			errmsg: "unknown command",
		},
		{
			title:  "flag without argument",
			args:   []string{"get", "-n", "-i"},
			errmsg: "flag needs an argument",
		},
		{
			title:  "invalid flag",
			args:   []string{"get", "-n", "--dummy"},
			errmsg: "flag provided but not defined",
		},
	}

	app.ErrorHandling = flag.ContinueOnError

	out := &strings.Builder{}
	app.Writer = out

	for _, c := range cases {
		out.Reset()
		opts.Clear()
		err := app.Parse(c.args)

		if len(c.errmsg) > 0 {
			if assert.Error(t, err, c.title) {
				assert.Contains(t, err.Error(), c.errmsg, c.title)
			}
		} else {
			if assert.NoError(t, err, c.title) {
				assert.Equal(t, c.expected, opts, c.title)
			}
		}

	}

}
