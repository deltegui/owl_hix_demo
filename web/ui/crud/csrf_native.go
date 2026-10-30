//go:build !(js && wasm)

package crud

func csrfToken() string { return "" }
