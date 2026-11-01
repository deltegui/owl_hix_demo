package coat

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/hxcoat"
	"github.com/deltegui/owl/v3"

	coatweb "owldemo/feature/coat/web"
	"owldemo/web/ui/bs"
)

// handleShowIndex writes the showcase with the server and loads the "coat"
// WASM, which hydrates it.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		return ctx.View(owl.HixView(func(owl.Ctx) hx.INode {
			return hx.El("html").Attribute("lang", ctx.GetCurrentLanguage()).Body(
				hx.El("head").Body(
					hx.El("meta").Attribute("charset", "utf-8"),
					hx.El("meta").Attribute("name", "viewport").Attribute("content", "width=device-width, initial-scale=1"),
					hx.El("title").Text("hxcoat · OWL demo"),
					hx.El("link").Attribute("rel", "stylesheet").Href(coat.CSS),
					hx.El("script").Src(coat.Tailwind)),
				hx.El("body").Body(
					bs.Island("coat").Body(coatweb.Island("")),
					hx.El("script").Src("/js/wasm_exec.js"),
					hx.Raw(`<script>const go = new Go();
WebAssembly.instantiateStreaming(fetch("/js/coat.wasm"), go.importObject).then((r) => go.run(r.instance));</script>`)))
		}))
	}
}
