// Package static has the files the browser loads: wasm_exec.js and the WASM
// binaries that `make build` writes in js/.
package static

import "embed"

//go:embed *
var Files embed.FS
