//go:build js && wasm

// Command lite is the WASM of the pages that are mostly text: it only carries
// the theme selector, which imports nothing heavy.
package main

import (
	"github.com/deltegui/hx"

	"owldemo/web/ui/island"
	"owldemo/web/ui/theme"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"theme": theme.Build,
	})
}
