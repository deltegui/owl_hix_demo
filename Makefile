.PHONY: build serve release check check-bulma check-coat check-bs test

# Two binaries: lite for the pages that are mostly text, full for the screens.
# They go in web/static/js, which the server embeds, so build before `go build`.
build:
	GOOS=js GOARCH=wasm go build -o web/static/js/lite.wasm ./client/lite
	GOOS=js GOARCH=wasm go build -o web/static/js/full.wasm ./client/full
	GOOS=js GOARCH=wasm go build -o web/static/js/forms.wasm ./client/forms
	GOOS=js GOARCH=wasm go build -o web/static/js/entries.wasm ./client/entries
	GOOS=js GOARCH=wasm go build -o web/static/js/bulma.wasm ./client/bulma
	GOOS=js GOARCH=wasm go build -o web/static/js/coat.wasm ./client/coat
	GOOS=js GOARCH=wasm go build -o web/static/js/bsdemo.wasm ./client/bsdemo
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/js/

serve: build
	go run ./cmd/owldemo

# The same two, optimized with TinyGo.
release:
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/lite.wasm ./client/lite
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/full.wasm ./client/full
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/forms.wasm ./client/forms
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/entries.wasm ./client/entries
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/bulma.wasm ./client/bulma
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/coat.wasm ./client/coat
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/bsdemo.wasm ./client/bsdemo
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" web/static/js/

test:
	go test -race ./...

# Needs node 22+ and chromium. Starts its own server and browser.
check: build
	node check.mjs

# The Bulma showcase (/bulma). Needs node 22+ and chromium; Bulma comes from a CDN.
check-bulma: build
	node check-bulma.mjs

# The Basecoat showcase (/coat). Needs node 22+ and chromium; Basecoat and Tailwind come from a CDN.
check-coat: build
	node check-coat.mjs

# The Bootstrap showcase (/bs). Needs node 22+ and chromium; Bootstrap comes from a CDN.
check-bs: build
	node check-bs.mjs
