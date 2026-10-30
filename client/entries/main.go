//go:build js && wasm

// Command entries is the WASM of the page of ledger entries: the maintenance
// and the theme selector.
package main

import (
	"github.com/deltegui/hx"

	entryweb "owldemo/feature/entry/web"
	"owldemo/web/ui/island"
	"owldemo/web/ui/theme"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"theme":   theme.Build,
		"entries": entryweb.Island,
	})
}
