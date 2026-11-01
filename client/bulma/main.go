//go:build js && wasm

// Command bulma is the WASM of the Bulma showcase.
package main

import (
	"github.com/deltegui/hx"

	bulmaweb "owldemo/feature/bulma/web"
	"owldemo/web/ui/island"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"bulma": bulmaweb.Island,
	})
}
