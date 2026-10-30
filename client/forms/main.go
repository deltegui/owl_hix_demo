//go:build js && wasm

// Command forms is the WASM of the pages with a form: the generic form and the
// theme selector.
package main

import (
	"github.com/deltegui/hx"

	formsweb "owldemo/feature/forms/web"
	"owldemo/web/ui/island"
	"owldemo/web/ui/theme"
)

func main() {
	island.Run(map[string]func(string) hx.INode{
		"theme": theme.Build,
		"form":  formsweb.Island,
	})
}
