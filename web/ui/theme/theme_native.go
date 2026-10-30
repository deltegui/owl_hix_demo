//go:build !(js && wasm)

package theme

// There is no page on a server: the theme is applied by the browser.

func saved() string { return "auto" }
func store(string)  {}
func apply(string)  {}
