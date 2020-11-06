package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/viper"
)

var allSources1 = []string{"source1", "source2", "source3", "sourceX"}

var yamlConfig1 = `
workers: 5
proxy: other
proxies:
  - proxy: tor
    url: socks5://127.0.0.1:9050
  - proxy: none
  - proxy: other
    url: https://127.0.0.1:7777
sources: 
  - source: source1
    proxy: none
  - source: source2
    proxy: tor
    workers: 2
  - source: source3
    workers: 3
isins:
  - isin: isin1
    name: Name of isin1 
    sources: [source1]
  - isin: isin2
    name: Name of isin2
    sources: 
      - source1
      - source2
`

// // set of string type
// type set map[string]struct{}

// func newSet(keys []string) set {
// 	s := set{}
// 	for _, k := range keys {
// 		s[k] = struct{}{}
// 	}
// 	return s
// }

// func (s set) has(key string) bool {
// 	_, ok := s[key]
// 	return ok
// }

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
func assertEqualString(t *testing.T, want, got string, msgAndArgs ...interface{}) {
	if got != want {
		msg := messageFromMsgAndArgs(msgAndArgs...)
		t.Errorf("%s: want %q, got %q", msg, want, got)
	}
}

func assertEqualInt(t *testing.T, want, got int, msgAndArgs ...interface{}) {
	if got != want {
		msg := messageFromMsgAndArgs(msgAndArgs...)
		t.Errorf("%s: want %d, got %d", msg, want, got)
	}
}

func assertEqualBool(t *testing.T, want, got bool, msgAndArgs ...interface{}) {
	if got != want {
		msg := messageFromMsgAndArgs(msgAndArgs...)
		t.Errorf("%s: want %v, got %v", msg, want, got)
	}
}

func assertError(t *testing.T, err error, substr string, msgAndArgs ...interface{}) {
	if len(substr) == 0 {
		if err == nil {
			msg := messageFromMsgAndArgs(msgAndArgs...)
			t.Errorf("%s: expected error", msg)
		}
		return
	}

	if err == nil || strings.Index(err.Error(), substr) < 0 {
		msg := messageFromMsgAndArgs(msgAndArgs...)
		t.Errorf("%s: expected error with substr %q, got: %q", msg, substr, err)
	}
}

func initViperConfig(config string) {
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // read in environment variables that match
	viper.ReadConfig(strings.NewReader(config))
}

func TestParseArgSource(t *testing.T) {
	testCases := []struct {
		input   string
		source  string
		workers int
		err     bool
	}{
		{
			input:  "source",
			source: "source",
		},
		{
			input:   "source:99",
			source:  "source",
			workers: 99,
		},
		{
			input:   "source/99",
			source:  "source",
			workers: 99,
		},
		{
			input:   "source#99",
			source:  "source",
			workers: 99,
		},
		{
			input: "source:",
			err:   true,
		},
		{
			input: "#99",
			err:   true,
		},
		{
			input: "source#nan",
			err:   true,
		},
	}
	for _, tc := range testCases {
		s, w, err := parseArgSource(tc.input, ":/#")
		if tc.err {
			assertError(t, err, "invalid source in args", tc.input)
		}
		if !tc.err {
			if err != nil {
				t.Errorf("Unxpected error for input %q: %v", tc.input, err)
			} else {
				// no error
				assertEqualString(t, tc.source, s, "input %q: source", tc.input)
				assertEqualInt(t, tc.workers, w, "input %q: workers", tc.input)
			}
		}
	}
}

func TestFullNotValidatedConfig(t *testing.T) {

	initViperConfig(yamlConfig1)

	args := &cmdGetArgs{
		Proxy:       "arg://proxy",
		PassedProxy: true,
		Isins:       []string{"isin1", "isinY"},
		Sources:     []string{"source1#101", "source2", "sourceY/12"},
	}

	cfg, err := getFullNotValidatedConfig(args, allSources1)
	if err != nil {
		t.Fatal(err)
	}

	if args.Sources == nil {
		s := cfg.Sources["sourceX"] // in all sources but not in config
		assertEqualBool(t, false, s.Disabled, "source[%q].disabled", s.Source)
	} else {
		swmap := map[string]int{}
		for _, sw := range args.Sources {
			s, w, _ := parseArgSource(sw, sepsSourceWorkers)
			swmap[s] = w
		}

		// check all args sources are found in cfg sources
		// check also the soruce.workers
		for s, w := range swmap {
			if source, ok := cfg.Sources[s]; !ok {
				t.Errorf("args source %q not found in cfg", s)
			} else if w != 0 {
				assertEqualInt(t, w, source.Workers, "source[%q].workers", source.Source)
			}
		}

		// check only sources in args are enabled
		for _, s := range cfg.Sources {
			_, ok := swmap[s.Source]
			assertEqualBool(t, !ok, s.Disabled, "source[%q].disabled", s.Source)
		}
	}

	if args.Isins != nil {
		// check all args isins are found in cfg isins
		for _, i := range args.Isins {
			if _, ok := cfg.Isins[i]; !ok {
				t.Errorf("args isin %q not found in cfg", i)
			}
		}
		// check only isins in args are enabled
		isinsArgsSet := newSet(args.Isins)
		for _, i := range cfg.Isins {
			ok := isinsArgsSet.has(i.Isin)
			assertEqualBool(t, !ok, i.Disabled, "isin[%q].disabled", i.Isin)
		}
	}

	// t.Log(cfg)
	// t.Fail()

}

func TestArgsProxy(t *testing.T) {
	testCases := []struct {
		title  string
		proxy  string
		passed bool
		want1  string
		want2  string
		want3  string
		wantX  string
	}{
		{
			title:  "passed",
			proxy:  "test://proxy",
			passed: true,
			want1:  "",
			want2:  "socks5://127.0.0.1:9050",
			want3:  "test://proxy",
			wantX:  "test://proxy",
		},
		{
			title:  "passed-ref",
			proxy:  "tor",
			passed: true,
			want1:  "",
			want2:  "socks5://127.0.0.1:9050",
			want3:  "socks5://127.0.0.1:9050",
			wantX:  "socks5://127.0.0.1:9050",
		},
		{
			title:  "passed-empty",
			proxy:  "",
			passed: true,
			want1:  "",
			want2:  "socks5://127.0.0.1:9050",
			want3:  "",
			wantX:  "",
		},
		{
			title:  "not-passed",
			passed: false,
			want1:  "",
			want2:  "socks5://127.0.0.1:9050",
			want3:  "https://127.0.0.1:7777",
			wantX:  "https://127.0.0.1:7777",
		},
	}

	initViperConfig(yamlConfig1)

	for j, tc := range testCases {
		if j < 0 {
			continue
		}

		args := &cmdGetArgs{
			Proxy:       tc.proxy,
			PassedProxy: tc.passed,
			Isins:       []string{"isinY"},
			Sources:     allSources1,
		}
		cfg, err := getConfig(args, allSources1)
		if err != nil {
			t.Fatal(err)
		}

		assertEqualString(t, tc.want1, cfg.Sources["source1"].Proxy, "%s: source1.proxy", tc.title)
		assertEqualString(t, tc.want2, cfg.Sources["source2"].Proxy, "%s: source1.proxy", tc.title)
		assertEqualString(t, tc.want3, cfg.Sources["source3"].Proxy, "%s: source3.proxy", tc.title)
		assertEqualString(t, tc.wantX, cfg.Sources["sourceX"].Proxy, "%s: sourceX.proxy", tc.title)
	}

}

func TestArgsWorkers(t *testing.T) {
	testCases := []struct {
		title   string
		workers int
		passed  bool
		want1   int
		want2   int
		want3   int
		wantX   int
	}{
		{
			title:   "passed",
			workers: 10,
			passed:  true,
			want1:   10,
			want2:   2,
			want3:   3,
			wantX:   10,
		},
		{
			title:  "not-passed",
			passed: false,
			want1:  5,
			want2:  2,
			want3:  3,
			wantX:  5,
		},
	}
	initViperConfig(yamlConfig1)

	for _, tc := range testCases {
		args := &cmdGetArgs{
			Workers:       tc.workers,
			PassedWorkers: tc.passed,
			Isins:         []string{"isinY"},
			Sources:       allSources1,
		}

		cfg, err := getConfig(args, allSources1)
		if err != nil {
			t.Fatal(err)
		}

		assertEqualInt(t, tc.want1, cfg.Sources["source1"].Workers, "%s: source1.workers", tc.title)
		assertEqualInt(t, tc.want2, cfg.Sources["source2"].Workers, "%s: source1.workers", tc.title)
		assertEqualInt(t, tc.want3, cfg.Sources["source3"].Workers, "%s: source3.workers", tc.title)
		assertEqualInt(t, tc.wantX, cfg.Sources["sourceX"].Workers, "%s: sourceX.workers", tc.title)
	}

}

func TestWorkersError(t *testing.T) {
	//
	// args.workers <= 0
	//
	initViperConfig(yamlConfig1)
	args := &cmdGetArgs{
		Workers:       0,
		PassedWorkers: true,
		Isins:         []string{"isinY"},
		Sources:       allSources1,
	}
	cfg, err := getConfig(args, allSources1)
	assertError(t, err, "workers must be greater than zero", "args.workers")

	//
	// cfg.workers = -1
	//
	args = &cmdGetArgs{
		Isins:   []string{"isinY"},
		Sources: allSources1,
	}
	initViperConfig("workers: -1")
	cfg, err = getConfig(args, allSources1)
	assertError(t, err, "workers must be greater than zero", "cfg.workers")

	//
	// cfg.workers = 0
	//
	initViperConfig("workers: 0")
	args = &cmdGetArgs{
		Isins:   []string{"isinY"},
		Sources: allSources1,
	}
	cfg, err = getConfig(args, allSources1)

	name := "source1"
	if cfg.Sources[name] == nil {
		t.Fatalf("source %q not found!", name)
	}
	assertEqualInt(t, defaultWorkers, cfg.Sources[name].Workers, "source[%q].workers", name)
}

func TestDefaults(t *testing.T) {
	flgs := getCmd.Flags()

	args := &cmdGetArgs{
		Isins:   []string{"isinY"},
		Sources: allSources1,
	}
	args.Workers, _ = flgs.GetInt("workers")
	args.Proxy, _ = flgs.GetString("proxy")

	initViperConfig("")

	cfg, err := getConfig(args, allSources1)
	if err != nil {
		t.Fatal(err)
	}

	name := "source1"
	if cfg.Sources[name] == nil {
		t.Fatalf("source %q not found!", name)
	}
	assertEqualInt(t, defaultWorkers, cfg.Sources[name].Workers, "source[%q].workers", name)
	assertEqualString(t, cfg.Sources[name].Proxy, "", "source[%q].proxy", name)
}

func TestSourcesFilter(t *testing.T) {
	initViperConfig(`
isins:
- isin: isin1
  sources: [source1]
- isin: isin2
  sources: [source1, source2]
`)
	cfg, err := getConfig(nil, allSources1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assertEqualInt(t, 2, len(cfg.Sources), "len(cfg.Sources)")
	for _, s := range []string{"source1", "source2"} {
		_, ok := cfg.Sources[s]
		assertEqualBool(t, true, ok, s)
	}
}

func TestSourcesUnknown(t *testing.T) {
	// 1. unknown source in config
	initViperConfig(`
isins:
- isin: isin1
  sources: [source1, sourceZ]
`)
	cfg, err := getConfig(nil, allSources1)
	assertError(t, err, "required source", "unknown source in config")

	// 2. unknown source in args
	args := &cmdGetArgs{
		Isins:   []string{"isinY"},
		Sources: []string{"sourceY"},
	}
	initViperConfig("")
	cfg, err = getConfig(args, allSources1)
	assertError(t, err, "required source", "unknown source in args")
	if t.Failed() {
		t.Log(cfg)
	}

	if t.Failed() {
		t.Log(cfg)
	}

}

func TestSourcesEmpty(t *testing.T) {
	initViperConfig(`
isins:
- isin: isin1
  sources: [source1]
sources:
- source: source1
  disabled: y
`)
	cfg, err := getConfig(nil, allSources1)
	assertError(t, err, "without enabled sources", "isin1")
	if t.Failed() {
		t.Log(cfg)
	}
}

func TestSourcesDisabled(t *testing.T) {
	initViperConfig(`
isins:
- isin: isin1
sources:
- source: source1
  disabled: y
`)
	cfg, err := getConfig(nil, allSources1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assertEqualInt(t, 3, len(cfg.Sources), "len(cfg.Sources)")
	for _, s := range []string{"source2", "source3", "sourceX"} {
		_, ok := cfg.Sources[s]
		assertEqualBool(t, true, ok, s)
	}
}

func TestSourceWorkers(t *testing.T) {
	initViperConfig(`
isins:
- isin: isin1
sources:
- source: source1
  workers: -1
`)
	_, err := getConfig(nil, allSources1)
	assertError(t, err, "workers must be greater than zero", "source1")
}

func TestSourceProxy(t *testing.T) {
	initViperConfig(`
isins:
- isin: isin1
sources:
- source: source1
  proxy: ::xxx
`)
	cfg, err := getConfig(nil, allSources1)
	assertError(t, err, "invalid proxy", "source1")
	if t.Failed() {
		t.Log(cfg)
	}
}

func TestKeyProxy(t *testing.T) {
	initViperConfig(`
isins:
- isin: isin1
proxies:
- url: htps://proxy
`)
	cfg, err := getConfig(nil, allSources1)
	assertError(t, err, "Invalid proxies: missing \"proxy\" key", "proxy")
	if t.Failed() {
		t.Log(cfg)
	}
}
func TestKeyIsin(t *testing.T) {
	initViperConfig(`
isins:
- sources: ["source1"]
`)
	cfg, err := getConfig(nil, allSources1)
	assertError(t, err, "Invalid isins: missing \"isin\" key", "isin")
	if t.Failed() {
		t.Log(cfg)
	}
}
func TestKeySource(t *testing.T) {
	initViperConfig(`
sources:
- proxy: https://proxy
`)
	cfg, err := getConfig(nil, allSources1)
	assertError(t, err, "Invalid sources: missing \"source\" key", "source")
	if t.Failed() {
		t.Log(cfg)
	}
}

func TestArgsDatabaase(t *testing.T) {
	db := "/home/user/config.toml"
	args := &cmdGetArgs{
		Database:       db,
		PassedDatabase: true,
		Isins:          []string{"isinY"},
		Sources:        allSources1,
	}
	initViperConfig("")
	cfg, err := getConfig(args, allSources1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assertEqualString(t, db, cfg.Database, "database")
}

func TestString(t *testing.T) {
	initViperConfig(yamlConfig1)
	cfg, err := getConfig(nil, allSources1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Log(cfg)
	// t.Fail()
}

func TestArgsInvalidSource(t *testing.T) {
	initViperConfig(yamlConfig1)
	args := &cmdGetArgs{
		Sources: []string{"source:nan"},
	}
	_, err := getConfig(args, allSources1)
	assertError(t, err, "invalid source in args")
}

func TestIsinSources(t *testing.T) {
	const isinY = "isinY"

	testCases := []struct {
		title           string
		args            *cmdGetArgs
		config          string
		allSources      []string
		expectedSources []string
	}{
		{
			title: "isinY in config: no source in args, no source in config",
			args:  nil,
			config: `isins: 
- isin: ` + isinY,
			allSources:      allSources1,
			expectedSources: allSources1,
		},
		{
			title: "isinY in args: no source in args, no source in config",
			args: &cmdGetArgs{
				Isins: []string{isinY},
			},
			config:          "",
			allSources:      allSources1,
			expectedSources: allSources1,
		},
		{
			title: "isinY in args: source in args overwrite source in config",
			args: &cmdGetArgs{
				Isins:   []string{isinY},
				Sources: []string{"source3", "sourceX"},
			},
			config: `
isins:
  - isin: isinY
    sources: [source1, source2]
`,
			allSources:      allSources1,
			expectedSources: []string{"sourceX", "source3"},
		},
		{
			title: "isinY in args: source in args are used even if disabled",
			args: &cmdGetArgs{
				Isins:   []string{isinY},
				Sources: []string{"source3", "sourceX"},
			},
			config: `
sources:
  - source: source3
    disabled: yes
`,
			allSources:      allSources1,
			expectedSources: []string{"sourceX", "source3"},
		},
		{
			title: "no source in args: sources disabled are not used",
			args:  nil,
			config: `
isins:
  - isin: isinY
sources:
  - source: source3
    disabled: yes
`,
			allSources:      allSources1,
			expectedSources: []string{"sourceX", "source1", "source2"},
		},
	}
	copts := cmp.Options{
		cmpopts.SortSlices(func(a, b string) bool {
			return a < b
		}),
	}

	for _, tc := range testCases {
		initViperConfig(tc.config)
		cfg, err := getConfig(tc.args, tc.allSources)
		if err != nil {
			t.Errorf("%s: error unexpected: %v", tc.title, err)
		} else if cfg.Isins[isinY] == nil {
			t.Errorf("%s: ???: %s not found in cfg", tc.title, isinY)
		} else if diff := cmp.Diff(tc.expectedSources, cfg.Isins[isinY].Sources, copts); diff != "" {
			t.Errorf("%s: mismatch (-want +got):\n%s", tc.title, diff)
		}

		if t.Failed() {
			t.Log(cfg)
		}

	}

}