.PHONY: build serve release check test

# Two binaries: lite for the pages that are mostly text, full for the screens.
# They go in web/static/js, which the server embeds, so build before `go build`.
build:
	GOOS=js GOARCH=wasm go build -o web/static/js/lite.wasm ./client/lite
	GOOS=js GOARCH=wasm go build -o web/static/js/full.wasm ./client/full
	GOOS=js GOARCH=wasm go build -o web/static/js/forms.wasm ./client/forms
	GOOS=js GOARCH=wasm go build -o web/static/js/entries.wasm ./client/entries
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/js/

serve: build
	go run ./cmd/owldemo

# The same two, optimized with TinyGo.
release:
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/lite.wasm ./client/lite
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/full.wasm ./client/full
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/forms.wasm ./client/forms
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/entries.wasm ./client/entries
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" web/static/js/

test:
	go test -race ./...

# Needs node 22+ and chromium. Starts its own server and browser.
check: build
	node check.mjs
