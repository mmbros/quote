// Copyright 2020 MMbros <server.mmbros@yandex.com>.
// Use of this source code is governed by Apache License.

/*
quote is a command line utility that retrieves stock/fund quotes from
various sources.

  quote <sub-command>

Available sub-commands are:

  get      Get the quotes of the specified isins
  sources  Show available sources
  tor      Checks if Tor network will be used


*/
package main

import "github.com/mmbros/quote/cmd"

func main() {
	cmd.Execute()
}
