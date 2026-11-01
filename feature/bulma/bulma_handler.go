package bulma

import (
	"github.com/deltegui/hx"
	hxbulma "github.com/deltegui/hxbulma"
	"github.com/deltegui/owl/v3"

	bulmaweb "owldemo/feature/bulma/web"
	"owldemo/web/ui/bs"
)

// handleShowIndex writes the showcase with the server and loads the "bulma"
// WASM, which hydrates it.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		return ctx.View(owl.HixView(func(owl.Ctx) hx.INode {
			return hx.El("html").Attribute("lang", ctx.GetCurrentLanguage()).Body(
				hx.El("head").Body(
					hx.El("meta").Attribute("charset", "utf-8"),
					hx.El("meta").Attribute("name", "viewport").Attribute("content", "width=device-width, initial-scale=1"),
					hx.El("title").Text("hxbulma · OWL demo"),
					hx.El("link").Attribute("rel", "stylesheet").Href(hxbulma.CSS)),
				hx.El("body").Body(
					bs.Island("bulma").Body(bulmaweb.Island("")),
					hx.El("script").Src("/js/wasm_exec.js"),
					hx.Raw(`<script>const go = new Go();
WebAssembly.instantiateStreaming(fetch("/js/bulma.wasm"), go.importObject).then((r) => go.run(r.instance));</script>`)))
		}))
	}
}
