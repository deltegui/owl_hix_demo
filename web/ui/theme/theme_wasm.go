//go:build js && wasm

package theme

import "syscall/js"

func saved() string {
	if v := js.Global().Get("localStorage").Call("getItem", "theme"); v.Truthy() {
		return v.String()
	}
	return "auto"
}

func store(choice string) { js.Global().Get("localStorage").Call("setItem", "theme", choice) }

func apply(choice string) {
	if choice == "auto" {
		choice = "light"
		if js.Global().Call("matchMedia", "(prefers-color-scheme: dark)").Get("matches").Bool() {
			choice = "dark"
		}
	}
	js.Global().Get("document").Get("documentElement").Call("setAttribute", "data-bs-theme", choice)
}
