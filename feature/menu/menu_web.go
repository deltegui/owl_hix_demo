package menu

import (
	"github.com/deltegui/owl/v3"

	"owldemo/web/layout"
)

const viewMenu string = "common/menu"

// getMenu is what the layout calls to draw the side menu, in the language of the
// request.
func getMenu(service menuService) func(owl.Ctx) []layout.MenuItem {
	return func(ctx owl.Ctx) []layout.MenuItem {
		loc := ctx.GetLocalizer(viewMenu)
		var items []layout.MenuItem
		for _, item := range service.items() {
			items = append(items, layout.MenuItem{Href: item.Href, Label: loc.Get(item.Key)})
		}
		return items
	}
}
