//go:build !(js && wasm)

package bs

import "github.com/deltegui/hx"

// PreventDefault does nothing on a server: there is no browser event.
func PreventDefault(hx.EventContext) {}
