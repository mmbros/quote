package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mmbros/quote/internal/quote"
	toml "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

const (
	defaultWorkers    = 1
	sepsSourceWorkers = ":/#"

	errmsgSourceNotAvailable        = "required source %q is not available"
	errmsgIsinWithoutEnabledSources = "isin %q without enabled sources"
	errmsgSourceWorkers             = "workers must be greater than zero (source %q has workers=%d)"
	errmsgWorkers                   = "workers must be greater than zero (workers=%d)"
	errmsgProxy                     = "invalid proxy: %s"
)

type sourceItem struct {
	Source   string `json:"source,omitempty"`
	Workers  int    `json:"workers,omitempty"`
	Proxy    string `json:"proxy,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type isinItem struct {
	Isin     string   `json:"isin,omitempty"`
	Name     string   `json:"name,omitempty"`
	Disabled bool     `json:"disabled,omitempty"`
	Sources  []string `json:"sources,omitempty"`
}

// Config is ...
type Config struct {
	Database string                 `json:"database,omitempty"`
	Workers  int                    `json:"workers,omitempty"`
	Proxy    string                 `json:"proxy,omitempty"`
	Proxies  map[string]string      `json:"proxies,omitempty"`
	Sources  map[string]*sourceItem `json:"sources,omitempty"`
	Isins    map[string]*isinItem   `json:"isins,omitempty"`
}

// String returns a json string representation of the object.
func jsonString(obj interface{}) string {
	// print config
	json, _ := json.MarshalIndent(obj, "", "  ")
	return string(json)
}

// parseArgSource gets the sourceWorkers string
// and returns the two components: source and workers.
// The components must be separated by one of the seps chars.
// If no separator char is found,
// retuns the input string as source and 0 as workers.
func parseArgSource(sourceWorkers, seps string) (source string, workers int, err error) {
	idx := strings.IndexAny(sourceWorkers, seps)
	if idx < 0 {
		source = sourceWorkers
	} else if idx == 0 || idx == len(sourceWorkers)-1 {
		goto labelReturnError
	} else {
		source = sourceWorkers[:idx]
		sw := sourceWorkers[idx+1:]
		workers, err = strconv.Atoi(sw)
		if err != nil {
			goto labelReturnError
		}
	}

	// check and normalize
	if workers == 0 {
		workers = defaultWorkers
	} else if workers < 0 {
		err = fmt.Errorf(errmsgSourceWorkers, source, workers)
	}
	return

labelReturnError:
	err = fmt.Errorf("invalid source in args: %q", sourceWorkers)
	return
}

// String returns a json string representation of the Config.
func (cfg *Config) String() string {
	return jsonString(cfg)
}

// resolveProxy returns the map value corresponding to the passed string.
// Returns the passed string if no corrisponding map item is found.
func (cfg *Config) resolveProxy(p string) string {
	if p == "" {
		p = cfg.Proxy
	}
	// if proxies map has a key corrisponding to p,
	// use the corrispondg map value
	// es. "tor"  -> "http://localhost:9090"
	if v, ok := cfg.Proxies[p]; ok {
		p = v
	}

	return p
}

// unmarshal parses data with given format to v object.
// Available formats are "json", "toml" or "yaml".
// Inn case format is not defined, all available format are tried.
func unmarshal(data []byte, v interface{}, dataFormat string) (err error) {

	type parseFunc func([]byte, interface{}) error

	parsers := map[string]parseFunc{
		"json": json.Unmarshal,
		"yaml": yaml.Unmarshal,
		"toml": toml.Unmarshal,
	}

	if dataFormat == "" {
		// try all known format
		for _, parse := range parsers {
			err = parse(data, v)
			if err == nil {
				return
			}
		}
		return fmt.Errorf("Unknown format")
	}

	// normalize format
	if dataFormat == "yml" {
		dataFormat = "yaml"
	}

	parse := parsers[dataFormat]
	if parse == nil {
		return fmt.Errorf("Unsupported format %q", dataFormat)
	}

	err = parse(data, v)

	return
}

// getFileFormat returns the (lowercase) data format base the followin criteria:
// 1. the explicitly passed format, or
// 2. the extension of the file without the leading "."
func getFileFormat(path string, format string) string {
	var s string

	if format != "" {
		s = format
	} else {
		s = filepath.Ext(path)
		if len(s) > 1 {
			s = s[1:]
		}
	}

	return strings.ToLower(s)
}

// addAllSources ensure that all available sources are listed in config,
// even if they are not passed in args or present in config file.
// The sources non already present are inserted with the passed disabled value.
//
// Even in case the sources are passed in args,
// must be called to returns specific error messages,
// i.e. to distiguish a disabled existing source from a not existing one.
func (cfg *Config) addAllSources(allSources []string, disabled bool) {
	for _, s := range allSources {
		source := cfg.Sources[s]
		if source == nil {
			// add new source
			cfg.Sources[s] = &sourceItem{
				Source:   s,
				Disabled: disabled,
			}
		}
	}
}

// normalizeVars complete the initialization of config varables.
// must be called after read and before merge.
func (cfg *Config) normalizeVars() {

	if cfg.Proxies == nil {
		cfg.Proxies = map[string]string{}
	}
	if cfg.Sources == nil {
		cfg.Sources = map[string]*sourceItem{}
	}
	if cfg.Isins == nil {
		cfg.Isins = map[string]*isinItem{}
	}

	// propagate map keys to values
	for k, v := range cfg.Isins {
		if v == nil {
			// isins:
			//   isin1:
			//     # nothing
			v = &isinItem{}
			cfg.Isins[k] = v
		}
		if v.Isin == "" {
			v.Isin = k
		}
	}
	for k, v := range cfg.Sources {
		if v == nil {
			v = &sourceItem{}
			cfg.Sources[k] = v
		}
		if v.Source == "" {
			v.Source = k
		}
	}
}

// merge updates config with passed argument and list of all sources.
//
// 1. ensures that all available sources are in config.Sources map;
// 2. sets Isin.Sources to allAvailableSources for isins that have Isin.Sources undefined;
// 3. updates config values with command line arguments, if defined.
//
// The output config has:
// - Isin and Source items disabled if not requested, enabled otherwise;
// - Isin.Sources not empty;
// - All available sourced existing in config.Sources
func (cfg *Config) merge(args *appArgs, allAvailableSources []string) error {

	// ensure all available source are in config
	disabled := (args != nil) && (len(args.sources) > 0)
	cfg.addAllSources(allAvailableSources, disabled)

	// sets Isin.Sources to allAvailableSources for isins that have Isin.Sources undefined.
	// If len(args.sources)>0 thi is not necessry, because it will be setted below.
	if (args == nil) || (len(args.sources) == 0) {
		for _, i := range cfg.Isins {
			if len(i.Sources) == 0 {
				i.Sources = allAvailableSources
			}
		}
	}

	if args == nil {
		return nil
	}

	// Database
	if args.database.Passed {
		cfg.Database = args.database.Value
	}

	// Workers
	if args.workers.Passed {
		w := args.workers.Value
		if w <= 0 {
			return fmt.Errorf(errmsgWorkers, w)
		}
		cfg.Workers = w
	} else {
		if cfg.Workers < 0 {
			return fmt.Errorf(errmsgWorkers, cfg.Workers)
		}
		if cfg.Workers == 0 {
			cfg.Workers = defaultWorkers
		}
	}

	// Proxy
	if args.proxy.Passed {
		cfg.Proxy = args.proxy.Value
	}

	// Isins
	//
	// If passed, only isins in args are getted
	// even if they are disabled in config!
	// Other isins in config are disabled.
	if len(args.isins) > 0 {
		// disable all the existing config isins
		for _, i := range cfg.Isins {
			i.Disabled = true
		}
		for _, i := range args.isins {
			item, ok := cfg.Isins[i]
			if ok {
				item.Disabled = false
			} else {
				item = &isinItem{
					Isin:    i,
					Sources: allAvailableSources,
				}
				cfg.Isins[i] = item
			}
		}
	}

	// Sources
	//
	// If sources are passed in args:
	// - only a source in args are used,
	//   even if they are disabled in config!
	//   Other sources in config are disabled.
	// - if in args the number of workers is specified for a source,
	//   the args workers value overwrite the config workers value.
	// - the isin.sources of the config file will be ignored:
	//   all the isins will use all and only the args.sources
	if len(args.sources) > 0 {
		var enabledSources []string
		// disable all the existing config sources
		for _, s := range cfg.Sources {
			s.Disabled = true
		}

		// needed to check sources arguments are unique (fix #11)
		// NOTE: no need to check isins are unique.
		mapArgsSourceToWorkers := map[string]int{}

		for _, sw := range args.sources {
			// split source from workers
			s, w, err := parseArgSource(sw, sepsSourceWorkers)
			if err != nil {
				return err
			}
			if precWorkers, ok := mapArgsSourceToWorkers[s]; ok {
				if precWorkers != w {
					return fmt.Errorf("duplicate source %q with different number of workers (%d and %d)", s, precWorkers, w)
				}
				continue // already inserted
			} else {
				mapArgsSourceToWorkers[s] = w
			}

			enabledSources = append(enabledSources, s)

			source, ok := cfg.Sources[s]
			if ok {
				// update existing source
				source.Disabled = false
				if w != 0 {
					source.Workers = w
				}
			} else {
				// add new source
				source = &sourceItem{
					Source:  s,
					Workers: w, // if 0, will be overwrite
				}
				cfg.Sources[s] = source
			}
		}
		// update Isins.sources with args sources
		for _, i := range cfg.Isins {
			i.Sources = enabledSources
		}
	}
	return nil
}

// reduce removes
// - isins disabled
// - sources not referenced (all disabled sources are NOT referenced)
func (cfg *Config) reduce(allSources []string) error {

	// set of (enabled) sources explicitly referenced by isins
	setRefEnabledSources := set{}

	// isins:
	for i, isin := range cfg.Isins {
		// remove disabled isin
		// see: https://stackoverflow.com/questions/23229975/is-it-safe-to-remove-selected-keys-from-map-within-a-range-loop
		if isin.Disabled {
			delete(cfg.Isins, i)
			continue
		}

		// filter and check isin sources
		isinEnabledSources := []string{}
		for _, s := range isin.Sources {
			source, ok := cfg.Sources[s]
			if !ok {
				// source not exists
				return fmt.Errorf(errmsgSourceNotAvailable, s)
			}
			if !source.Disabled {
				isinEnabledSources = append(isinEnabledSources, s)
				setRefEnabledSources.add(s)
			}
		}
		if len(isinEnabledSources) == 0 {
			// no sources
			return fmt.Errorf(errmsgIsinWithoutEnabledSources, i)
		}
		// update with filtered sources
		isin.Sources = isinEnabledSources
	}

	// remove not referenced sources
	// NOTE: all disabled sources are NOT referenced
	for s := range cfg.Sources {
		if !setRefEnabledSources.has(s) {
			delete(cfg.Sources, s)
		}
	}

	return nil
}

func (cfg *Config) check(allSources []string) error {

	setOfAllSources := newSet(allSources)

	// check proxy and workers of each referenced source
	for s, source := range cfg.Sources {
		// check source is available
		if !setOfAllSources.has(s) {
			return fmt.Errorf(errmsgSourceNotAvailable, s)
		}

		// check workers
		if source.Workers < 0 {
			return fmt.Errorf(errmsgSourceWorkers, source.Source, source.Workers)
		}
		if source.Workers == 0 {
			source.Workers = cfg.Workers
		}

		// proxy
		proxyURL := cfg.resolveProxy(source.Proxy)
		if proxyURL != "" {
			if _, err := url.Parse(proxyURL); err != nil {
				return fmt.Errorf(errmsgProxy, proxyURL)
			}
		}
		source.Proxy = proxyURL
	}

	return nil
}

// SourceIsinsList ...
// If no sources, returns a list with zero items (it does not returns nil).
// NOTE: it assumes all isins and sources are enabled
func (cfg *Config) SourceIsinsList() []*quote.SourceIsins {

	// build a map from (enabled) source to (enabled) isins
	sources := map[string][]string{}
	for _, i := range cfg.Isins {
		// skip disabled isins
		// if i.Disabled {
		// 	continue
		// }

		for _, s := range i.Sources {
			// skip disabled sources
			// if cfg.Sources[s].Disabled {
			// 	continue
			// }

			a := sources[s]
			if a == nil {
				a = []string{i.Isin}
			} else {
				a = append(a, i.Isin)
			}
			sources[s] = a
		}
	}

	sis := make([]*quote.SourceIsins, 0, len(sources))
	for s, isins := range sources {
		src := cfg.Sources[s]

		si := &quote.SourceIsins{
			Source:  s,
			Proxy:   src.Proxy,
			Workers: src.Workers,
			Isins:   isins,
		}
		sis = append(sis, si)
	}
	return sis
}

func (cfg *Config) auxGetConfig(data []byte, args *appArgs, allSources []string) error {
	var err error

	// 1. unmarshall data
	if (data != nil) && (len(data) > 0) {
		var path, format string
		if args != nil {
			path = args.config.Value
			format = args.configType.Value
		}
		format = getFileFormat(path, format)
		err = unmarshal(data, cfg, format)
	}

	// 2. normalize config variables
	cfg.normalizeVars()

	// 3. merge command line arguments in config
	if err == nil {
		err = cfg.merge(args, allSources)
	}

	// 4. remove unused isins and sources
	if err == nil {
		err = cfg.reduce(allSources)
	}

	// 5. check config
	if err == nil {
		err = cfg.check(allSources)
	}

	return err
}

// GetConfig ...
func GetConfig(args *appArgs, allSources []string) (*Config, error) {
	var (
		err  error
		data []byte
	)
	cfg := &Config{}

	// 1. read config file, if defined
	if (args != nil) && (args.config.Value != "") {
		data, err = ioutil.ReadFile(args.config.Value)
		if err != nil {
			return nil, err
		}
	}

	err = cfg.auxGetConfig(data, args, allSources)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
