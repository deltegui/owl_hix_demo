//go:build js && wasm

// Command coat is the WASM of the Basecoat showcase.
package main

import (
	"github.com/deltegui/hx"

	coatweb "owldemo/feature/coat/web"
	"owldemo/web/ui/island"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"coat": coatweb.Island,
	})
}
