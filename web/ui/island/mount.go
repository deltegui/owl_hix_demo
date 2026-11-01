//go:build js && wasm

// Package island is what every WASM entry point does: hydrate the islands that
// the page has. The mechanism is OWL's (owl/island); this is the entry point for
// the islands that are not defined with island.Define yet and still come as a
// registry of functions over the text of the state.
package island

import (
	"github.com/deltegui/hx"
	owl "github.com/deltegui/owl/v3/island"
)

// Run hydrates the islands of the registry and the ones given, and keeps the
// program alive.
func Run(registry map[string]func(state string) hx.INode, defined ...owl.Island) {
	islands := defined
	for name, build := range registry {
		islands = append(islands, owl.Func(name, build))
	}
	owl.Run(islands...)
}
