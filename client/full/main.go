//go:build js && wasm

// Command full is the WASM of the pages with screens: it carries the islands of
// every feature, and with them the JSON code.
package main

import (
	"github.com/deltegui/hx"

	contactweb "owldemo/feature/contact/web"
	productweb "owldemo/feature/product/web"
	"owldemo/web/ui/island"
	"owldemo/web/ui/theme"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"theme":    theme.Build,
		"contacts": contactweb.ListIsland,
	}, productweb.StatsIsland, productweb.ListIsland)
}
