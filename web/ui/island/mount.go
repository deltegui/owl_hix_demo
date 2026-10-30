//go:build js && wasm

// Package island is what every WASM entry point does: hydrate the islands that
// the page has. Each binary calls it with the islands it carries.
package island

import (
	"log"
	"syscall/js"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/wasm"

	"owldemo/web/ui/bs"
)

// Run hydrates each island of the registry that the page has, and then keeps the
// program alive. The server rendered the island inside its element and put the
// state it used in a script element next to it; the island builds from that
// same state, so what it builds is what is already there and nothing is redrawn.
// An island in the page that the registry does not have is a page that loaded
// the wrong binary, and it is reported.
func Run(registry map[string]func(state string) hx.INode) {
	document := js.Global().Get("document")
	found := document.Call("querySelectorAll", "[data-island]")
	for i := 0; i < found.Length(); i++ {
		if name := found.Index(i).Get("dataset").Get("island").String(); registry[name] == nil {
			log.Printf("the page has the island %q, and this binary does not carry it", name)
		}
	}
	for name, build := range registry {
		if document.Call("getElementById", bs.IslandID(name)).IsNull() {
			continue // this page does not have it
		}
		state := wasm.ReadState(bs.StateID(name))
		wasm.Hydrate(bs.IslandID(name), func() hx.INode { return build(state) })
	}
	select {}
}
