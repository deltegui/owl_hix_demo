package bsdemo

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/hxbs"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"
	"github.com/deltegui/owl/v3/island"

	bsweb "owldemo/feature/bsdemo/web"
)

// handleShowIndex writes the showcase with the server and loads the "bsdemo"
// WASM, which hydrates it.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		return ctx.View(owl.HixView(func(owl.Ctx) hx.INode {
			page := island.Instance{Island: island.Func("bsdemo", bsweb.Island)}
			return hxui.App(ctx, hxui.AppOptions{
				Title:   "hxbs · OWL demo",
				Head:    []hx.INode{hxui.Stylesheet(bs.CSS)},
				Bundle:  "bsdemo",
				Islands: []island.Instance{page},
			}, page.Node())
		}))
	}
}
