//go:build js && wasm

package bs

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/hx/wasm"
)

// PreventDefault stops the browser from doing what it does by itself with the
// event, for example sending a form. On a server there is no event: it does nothing.
func PreventDefault(ctx hx.EventContext) { wasm.PreventDefault(ctx) }
