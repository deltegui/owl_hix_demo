package bulma

import (
	"github.com/deltegui/hx"
	hxbulma "github.com/deltegui/hxbulma"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"
	"github.com/deltegui/owl/v3/island"

	bulmaweb "owldemo/feature/bulma/web"
)

// handleShowIndex writes the showcase with the server and loads the "bulma"
// WASM, which hydrates it.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		return ctx.View(owl.HixView(func(owl.Ctx) hx.INode {
			page := island.Instance{Island: island.Func("bulma", bulmaweb.Island)}
			return hxui.App(ctx, hxui.AppOptions{
				Title:   "hxbulma · OWL demo",
				Head:    []hx.INode{hxui.Stylesheet(hxbulma.CSS)},
				Bundle:  "bulma",
				Islands: []island.Instance{page},
			}, page.Node())
		}))
	}
}
