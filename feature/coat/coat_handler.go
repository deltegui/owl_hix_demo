package coat

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/hxcoat"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"
	"github.com/deltegui/owl/v3/island"

	coatweb "owldemo/feature/coat/web"
)

// handleShowIndex writes the showcase with the server and loads the "coat"
// WASM, which hydrates it.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		return ctx.View(owl.HixView(func(owl.Ctx) hx.INode {
			page := island.Instance{Island: island.Func("coat", coatweb.Island)}
			return hxui.App(ctx, hxui.AppOptions{
				Title:   "hxcoat · OWL demo",
				Head:    []hx.INode{hxui.Stylesheet(coat.CSS), hx.El("script").Src(coat.Tailwind)},
				Bundle:  "coat",
				Islands: []island.Instance{page},
			}, page.Node())
		}))
	}
}
