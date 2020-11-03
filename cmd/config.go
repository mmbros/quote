package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type sourceItem struct {
	Source   string
	Workers  int
	Proxy    string
	Disabled bool
}
type isinItem struct {
	Isin     string
	Name     string
	Disabled bool
	Sources  []string
}

// Config is ...
type Config struct {
	Database string
	Workers  int
	Proxy    string
	Proxies  map[string]string
	Sources  map[string]*sourceItem
	Isins    map[string]*isinItem
}

type cmdGetArgs struct {
	Database       string
	Workers        int
	Proxy          string
	Sources        []string
	Isins          []string
	DryRun         bool
	passedDatabase bool
	passedWorkers  bool
	passedProxy    bool
}

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

// String returns a json string representation of the config.
func (cfg *Config) String() string {
	// print config
	json, _ := json.MarshalIndent(cfg, "", "  ")
	return string(json)
}

// readSources init the proxies from config file.
// Error is returned only if the key of a proxy is invalid.
func (cfg *Config) readProxies(vip *viper.Viper) error {
	// https://lornajane.net/posts/2020/accessing-nested-config-with-viper

	type CfgProxy struct {
		Proxy string
		URL   string
	}
	var cfgProxies []*CfgProxy

	proxies := map[string]string{}

	err := vip.UnmarshalKey("proxies", &cfgProxies)
	if err != nil {
		return err
	}
	for _, p := range cfgProxies {
		if p.Proxy == "" {
			return fmt.Errorf("Invalid proxies: missing \"proxy\" key")
		}
		proxies[p.Proxy] = p.URL
	}

	cfg.Proxies = proxies
	return nil
}

// readSources init the sources from config file.
// Error is returned only if the key of a source is invalid.
func (cfg *Config) readSources(vip *viper.Viper) error {
	var cfgSources []*sourceItem
	sources := map[string]*sourceItem{}

	if err := vip.UnmarshalKey("sources", &cfgSources); err != nil {
		return err
	}
	for _, item := range cfgSources {
		if item.Source == "" {
			return fmt.Errorf("Invalid sources: missing \"source\" key")
		}
		sources[item.Source] = item
	}

	cfg.Sources = sources
	return nil
}

// readIsins init the isins from config file.
// Error is returned only if the key of an isin is invalid.
func (cfg *Config) readIsins(vip *viper.Viper) error {
	var cfgIsins []*isinItem
	isins := map[string]*isinItem{}

	if err := vip.UnmarshalKey("isins", &cfgIsins); err != nil {
		return err
	}
	for _, item := range cfgIsins {
		if item.Isin == "" {
			return fmt.Errorf("Invalid isins: missing \"isin\" key")
		}
		isins[item.Isin] = item
	}

	cfg.Isins = isins
	return nil
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
	return

labelReturnError:
	err = fmt.Errorf("invalid source in args: %q", sourceWorkers)
	return
}

// readConfig init the config from config file.
// Error is returned only if a key of proxy, source or isin is invalid.
func (cfg *Config) readConfig(vip *viper.Viper) error {
	cfg.Database = vip.GetString("database")
	cfg.Proxy = vip.GetString("proxy")
	cfg.Workers = vip.GetInt("workers")

	if err := cfg.readProxies(vip); err != nil {
		return err
	}
	if err := cfg.readSources(vip); err != nil {
		return err
	}
	if err := cfg.readIsins(vip); err != nil {
		return err
	}

	return nil
}

// mergeArgs updates the config values with the passed arguments.
// error is returned in in case of parseArgSource error
func (cfg *Config) mergeArgs(args *cmdGetArgs) error {

	if args == nil {
		return nil
	}

	// workers
	if args.passedWorkers {
		if args.Workers <= 0 {
			return fmt.Errorf("workers must be greater than 0 (found %d)", cfg.Workers)
		}
		cfg.Workers = args.Workers
	}

	// proxy
	if args.passedProxy {
		cfg.Proxy = args.Proxy
	}

	// database
	if args.passedDatabase {
		cfg.Database = args.Database
	}

	// isins
	//
	// If passed, only isins in args are getted
	// even if they are disabled in config!
	// Other isins in config are disabled.
	if len(args.Isins) > 0 {
		// disable all the existing config isins
		for _, i := range cfg.Isins {
			i.Disabled = true
		}
		for _, i := range args.Isins {
			item, ok := cfg.Isins[i]
			if ok {
				item.Disabled = false
			} else {
				item = &isinItem{
					Isin: i,
				}
				cfg.Isins[i] = item
			}
		}
	}

	// sources
	//
	// If passed only a source in args are used,
	// even if they are disabled in config!
	// Other sources in config are disabled.
	// If in args the number of workers is specified for a source,
	// the args workers value overwrite the config workers value.
	if len(args.Sources) > 0 {
		// disable all the existing config sources
		for _, s := range cfg.Sources {
			s.Disabled = true
		}
		for _, sw := range args.Sources {
			s, w, err := parseArgSource(sw, sepsSourceWorkers)
			if err != nil {
				return err
			}
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
					Workers: w,
				}
				cfg.Sources[s] = source
			}
		}
	}
	return nil
}

// addAllSources ensure that alla available sources are listed in config,
// even if they are not passed in args or present in config file.
// The sources non already present are inserted with the passed disabled value.
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

func getFullNotValidatedConfig(args *cmdGetArgs, allSources []string) (*Config, error) {
	vip := viper.GetViper()
	cfg := &Config{}

	if err := cfg.readConfig(vip); err != nil {
		return nil, err
	}
	if err := cfg.mergeArgs(args); err != nil {
		return nil, err
	}
	disabled := (args != nil) && (len(args.Sources) > 0)
	cfg.addAllSources(allSources, disabled)

	// at this point the config have all the components,
	// but may have invalid source, proxy, workers ...

	// TODO:
	// 1. filter (consider only used isin, source and proxy)
	// 2. check
	// 3. insert defaults and references

	// cfg.filterDisabled()

	return cfg, nil
}

// checkAndSimplify the config with:
// - filter (consider only used isin, source and proxy)
// - check
// - insert defaults and references
func (cfg *Config) checkAndSimplify() error {

	// check workers

	// postponed
	// if cfg.Workers < 0 {
	// 	return fmt.Errorf("workers must be greater than zero (found %d)", cfg.Workers)
	// }
	if cfg.Workers == 0 {
		cfg.Workers = defaultWorkers
	}

	// allEnabledSources is true if at least one (enabled) isin
	// can use all (enable) sources
	flagAllEnabledSources := false

	// map of (enabled) sources explicitly referenced by isins
	refSources := map[string]*sourceItem{}

	// check anf filter isins
	for i, isin := range cfg.Isins {
		// remove disabled ISIN
		// see: https://stackoverflow.com/questions/23229975/is-it-safe-to-remove-selected-keys-from-map-within-a-range-loop
		if isin.Disabled {
			delete(cfg.Isins, i)
			continue
		}

		// check isin sources
		if len(isin.Sources) == 0 {
			flagAllEnabledSources = true
		} else {
			allEnabledSources := []string{}
			for _, s := range isin.Sources {
				source, ok := cfg.Sources[s]
				if !ok {
					// source not exists
					return fmt.Errorf("isin %q with unknown source %q", i, s)
				}
				if !source.Disabled {
					allEnabledSources = append(allEnabledSources, s)
					refSources[s] = source
				}
			}
			if len(allEnabledSources) == 0 {
				// no filtered sources
				return fmt.Errorf("isin %q without enabled sources", i)
			}
			// update with filtered sources
			isin.Sources = allEnabledSources
		}
	}

	// filter Sources
	if flagAllEnabledSources {
		// list of all (enabled) sources
		allEnabledSources := []string{}
		for s, source := range cfg.Sources {
			if source.Disabled {
				delete(cfg.Sources, s)
			} else {
				allEnabledSources = append(allEnabledSources, s)
			}
		}

		// update isin sources of isins with no sources list
		for _, isin := range cfg.Isins {
			if len(isin.Sources) == 0 {
				isin.Sources = allEnabledSources
			}
		}

	} else {
		// every isin has an explicit source list
		cfg.Sources = refSources
	}

	// set proxy and workers of each source
	for _, source := range cfg.Sources {

		// workers
		if source.Workers < 0 {
			return fmt.Errorf("workers must be greater than zero (source %q has workers=%d)",
				source.Source, source.Workers)
		}
		if source.Workers == 0 {
			if cfg.Workers < 0 {
				return fmt.Errorf("workers must be greater than zero (workers=%d)",
					cfg.Workers)
			}
			source.Workers = cfg.Workers
		}

		// proxy
		if source.Proxy == "" {
			source.Proxy = cfg.Proxy
		}
		if p, ok := cfg.Proxies[source.Proxy]; ok {
			// if the proxy match a key of the proxy map,
			// set the proxy to the corresponding value
			source.Proxy = p
		}
		if source.Proxy != "" {
			if _, err := url.Parse(source.Proxy); err != nil {
				return fmt.Errorf("invalid proxy: %s", source.Proxy)
			}
		}
	}

	return nil
}

func getConfig(args *cmdGetArgs, allSources []string) (*Config, error) {
	cfg, err := getFullNotValidatedConfig(args, allSources)
	if err == nil {
		err = cfg.checkAndSimplify()
	}
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
