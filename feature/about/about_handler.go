package about

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"

	"owldemo/web/layout"
	"owldemo/web/ui/bs"
)

const viewAboutIndex string = "about/about_index"

// handleShowIndex is a page that is only HTML, built with Hix on the server. It
// has no islands: it loads the small WASM, which only carries the theme selector.
func handleShowIndex() owl.Handler {
	return func(ctx owl.Ctx) error {
		t := ctx.GetLocalizer(viewAboutIndex)
		return layout.Render(ctx, layout.Page{
			Title:  t.Get("Title"),
			Bundle: "lite",
			Content: hx.Div().Class("container-fluid").Body(
				hx.H1().Text(t.Get("Title")),
				hx.P().Class("lead").Text(t.Get("Subtitle")),
				bs.Card("", hx.P().Text(t.Get("Body1")), hx.P().Text(t.Get("Body2")), hx.P().Class("mb-0", "text-muted").Text(t.Get("Features")))),
		})
	}
}
