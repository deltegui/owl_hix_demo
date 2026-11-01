//go:build js && wasm

// Command bsdemo is the WASM of the Bootstrap showcase.
package main

import (
	"github.com/deltegui/hx"

	bsweb "owldemo/feature/bsdemo/web"
	"owldemo/web/ui/island"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"bsdemo": bsweb.Island,
	})
}
