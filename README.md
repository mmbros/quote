# quote
Get stock/fund quotes from various sources


`quote` is a command line utility that retrieves stock/fund quotes from various sources.

The stock/fund secureties are identified by their International Securities Identification Number (ISIN). 

Each quote request is retrieved concurrently from all the sources available for that stock/fund. For each isin, the first success request is returned, and the remaining requests are cancelled.

See `quote sources` for a list of the available sources.

The number of workers of each source represents the number of concurrent requests that can be executes for that particular source.

A configuration file, not mandatory, can be used to save the parameters and fine tuning the retrieve of the quotes.

*Example*

    quote get -i isin1,isin2 -s sourceA/4,sourceB, -s sourceC --workers 2

It retrieves the quotes of 2 isins from 3 sources: A with 4 workers, B and C with 2 workers each.

## Commands

### `quote` command

    Usage:
      quote [command]
    
    Available Commands:
      get         Get the quotes of the specified isins
      help        Help about any command
      sources     Show available sources
      tor-check   Checks if Tor network will be used
    
    Flags:
          --config string     config file (default is $HOME/.    quote.yaml)
          --database string   quote sqlite3 database
      -h, --help              help for quote
          --proxy string      default proxy

### `quote get` subcommand
 
Get the quotes of the specified isins from the sources.
If source options are not specified, all the available sources for the isin are used.
See `quote sources` for a list of the available sources.

    Usage:
      quote get [flags]
    
    Examples:
        quote get -i isin1,isin2 -s sourceA/4,sourceB, -s     sourceC --workers 2
      retrieves 2 isins from 3 sources: A with 4 workers, B and     C with 2 workers each.
    
    Flags:
      -n, --dry-run           perform a trial run with no     request/updates made
      -h, --help              help for get
      -i, --isins strings     list of isins to get the quotes
      -s, --sources strings   list of sources to get the quotes     from
      -w, --workers int       number of workers (default 1)
    
    Global Flags:
          --config string     config file (default is $HOME/.    quote.yaml)
          --database string   quote sqlite3 database
          --proxy string      default proxy


### `quote sources` subcommand

Show available sources.

*Example:*

    $ quote sources
    > [cryptonatorcom-EUR fondidocit fundsquarenet     morningstarit]

### `quote tor` subcommand

Checks if the quotes are retrieved through the Tor network.

To use the Tor network the proxy must be defined through:
  1. `--proxy` or `-p` argument parameter
  2. `proxy` config file parameter
  3. `HTTP_PROXY`, `HTTPS_PROXY` and `NOPROXY` enviroment variables.


## Config file

