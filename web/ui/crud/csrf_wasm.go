//go:build js && wasm

package crud

import "syscall/js"

// csrfToken is the token the server put in the page (see layout), which it asks
// for in every request that changes something.
func csrfToken() string {
	meta := js.Global().Get("document").Call("querySelector", `meta[name="csrf-token"]`)
	if meta.IsNull() {
		return ""
	}
	return meta.Call("getAttribute", "content").String()
}
