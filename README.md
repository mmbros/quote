# quote
Get stock/fund quotes from various sources.

[![Build Status](https://travis-ci.com/mmbros/quote.svg?branch=main)](https://travis-ci.com/mmbros/quote)

`quote` is a command line utility that retrieves stock/fund quotes from
various sources.

The stock/fund securities are identified by their International Securities
Identification Number (ISIN). 
Cryptocurrencies are identified by their currency code.

Each quote request is retrieved concurrently from all the sources available
for that stock/fund. For each isin, the first success request is returned,
and the remaining requests are cancelled.

See `quote sources` for a list of the available sources.

The number of workers of each source represents the number of concurrent
requests that can be executed for that specific source.

A configuration file, not mandatory, can be used to save the parameters and
fine tuning the retrieve of the quotes.

*Example*

    quote get -i isin1,isin2 -s sourceA/4,sourceB, -s sourceC --workers 2

It retrieves the quotes of 2 isins from 3 sources: A with 4 workers,
B and C with 2 workers each.

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
          --config string     config file (default is $HOME/.quote.yaml)
          --database string   quote sqlite3 database
      -h, --help              help for quote
          --proxy string      default proxy

### `quote get` sub-command
 
Get the quotes of the specified isins from the sources.
If source options are not specified, all the available sources for
the isin are used.

See `quote sources` for a list of the available sources.

    Usage:
      quote get [flags]
    
    Examples:
        quote get -i isin1,isin2 -s sourceA/4,sourceB, -s sourceC --workers 2
      retrieves 2 isins from 3 sources: A with 4 workers, B and C with 2 workers each.
    
    Flags:
      -n, --dry-run           perform a trial run with no request/updates made
      -h, --help              help for get
      -i, --isins strings     list of isins to get the quotes
      -s, --sources strings   list of sources to get the quotes from
      -w, --workers int       number of workers (default 1)
    
    Global Flags:
          --config string     config file (default is $HOME/.quote.yaml)
          --database string   quote sqlite3 database
          --proxy string      default proxy


### `quote sources` sub-command

Show available sources.

*Example:*

    $ quote sources
    > [cryptonatorcom-EUR fondidocit fundsquarenet morningstarit]

### `quote tor` sub-command

Checks if the quotes are retrieved through the Tor network.

To use the Tor network the proxy can be defined through:
  1. `--proxy` or `-p` argument parameter
  2. `proxy` config file parameter
  3. `HTTP_PROXY`, `HTTPS_PROXY` and `NOPROXY` enviroment variables.


## Configuration file

The quote configuration file can be written in `toml`, `yaml` or `json` format.


### `config`

|param   |type  |description|
|--------|------|-|
|database|string|path of the sqlite3 database where the quotes are saved. If setted, the database is created if not exists.|
|workers |int   |Default number of workers. Used if param `workers` is missing for sources without specific `workers` value.|
|proxy   |string|Default proxy. Used if param `proxy` is missing for sources without specific `proxy` value.|
|proxies |array |List of proxies to be used. See below for proxy fields.|
|isins   |array |List of isins to be retrieved. See below for isin fields.|
|sources |array |List of sources. See below for source fields.|

### `proxies`
List of proxies to be used.

|param   |type  |description|
|--------|------|-|
|proxy   |string|Mandatory name of the proxy.| 
|url     |string|URL of the proxy.|

### `isins`
List of isins to be retrieved.

|param   |type  |description|
|--------|------|-|
|isin    |string|Mandatory ID of the fund/stock.| 
|name    |string|Name of the fund/stock. Only for documentation porpouses; it's not used in the retrieval of the quote.|
|sources |array |List of the sources to be used to get the quote of the isin. If missing, all the (enabled) available sources are used.|
|disabled|bool  |If disabled, the isin is not retrieved.|


In case `--isin` argument is setted in the command line: 

- only the isins passed in the command line are retrieved, 
  even if they don't exists or are disabled in the config file;
- if the `sources` param is setted in the config file for an isin passed as an argument,
  only those sources are used (if not disabled) to retrieve the quote,
  even if the isin is disabled in the config file.


### `sources`
List of sources in the configuration files.

The complete list of available sources can be getted with `quote sources`.
 
 Each source configuration section can have the following fields:

|param   |type  |description|
|--------|------|-|
|source  |string|Mandatory name of the source.| 
|workers |int   |Number of workers.|
|proxy   |string|Proxy url or proxy name to be used.|
|disabled|bool  |If disabled, the source is not used.|

In case `--source` argument is passed in the command line: 

- only the sources passed in the command line are used,
  even if they don't exists or are disabled in the config file;


### Example
Configuration file in `yaml` format.

    database: /home/user/quote.sqlite3
    
    workers: 2
    
    proxy: tor
    
    proxies:
      - proxy: tor
        url: socks5://127.0.0.1:9050
      - proxy: none
    
    isins:
      - isin: BTC
        name: Bitcoin
        sources: [crypto1]
    
      - isin: ETH
        name: Ethereum
        sources: [crypto1]
    
      - isin: LU0000000001
        name: Global Fund USD 2020
        sources: [source1, source2, source3]
    
      - isin: LU0000000002
        name: China Dynamic Fund Acc EUR 
        sources: [source1, source2, source3]
    
      - isin: EU0000000003
        name: Growth Europe EUR III
        sources: [source1, source2, source3]
        disabled: true
    
    sources:
      - source: crypto1
        proxy: none
        workers: 1
    
      - source: source1
        workers: 3
    
      - source: source3
        disabled: true
    
