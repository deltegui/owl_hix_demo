package bsdemo

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3"

	bsweb "owldemo/feature/bsdemo/web"
	ui "owldemo/web/ui/bs"
)

// handleShowIndex writes the showcase with the server and loads the "bsdemo"
// WASM, which hydrates it.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		return ctx.View(owl.HixView(func(owl.Ctx) hx.INode {
			return hx.El("html").Attribute("lang", ctx.GetCurrentLanguage()).Body(
				hx.El("head").Body(
					hx.El("meta").Attribute("charset", "utf-8"),
					hx.El("meta").Attribute("name", "viewport").Attribute("content", "width=device-width, initial-scale=1"),
					hx.El("title").Text("hxbs · OWL demo"),
					hx.El("link").Attribute("rel", "stylesheet").Href(bs.CSS)),
				hx.El("body").Body(
					ui.Island("bsdemo").Body(bsweb.Island("")),
					hx.El("script").Src("/js/wasm_exec.js"),
					hx.Raw(`<script>const go = new Go();
WebAssembly.instantiateStreaming(fetch("/js/bsdemo.wasm"), go.importObject).then((r) => go.run(r.instance));</script>`)))
		}))
	}
}
