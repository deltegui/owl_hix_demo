package productweb

import (
	"fmt"

	"github.com/deltegui/hx"

	"owldemo/web/ui/crud"
)

// Version goes up every time a product changes, in any island of the page. It
// does not hold data (the server does): it says that whoever has read the
// products should read them again, so the chart follows the list.
var Version = hx.Signal(0)

func Changed() { Version.Update(func(n int) int { return n + 1 }) }

// Screen is the list of products. state.Page is the first page, which the server
// rendered; if it is nil the screen loads it.
func Screen(state crud.State[Product]) hx.INode {
	t := TextsFor(state.Lang)
	return crud.Screen(crud.Config[Product]{
		Texts:   t,
		API:     "/api/products",
		Initial: state.Page,
		Changed: Changed,
		ID:      func(p Product) int { return p.ID },
		Name:    func(p Product) string { return p.Name },
		Active:  func(p Product) bool { return p.Active },
		Columns: []crud.Column[Product]{
			crud.TextCol("#", func(p Product) string { return fmt.Sprint(p.ID) }),
			crud.TextCol("Name", func(p Product) string { return p.Name }),
			crud.TextCol("Category", func(p Product) string { return t.Or("Opt."+p.Category, p.Category) }),
			crud.TextCol("Price", func(p Product) string { return fmt.Sprintf("%.2f", p.Price) }),
			crud.ActiveCol("Status", func(p Product) bool { return p.Active }),
		},
		Filters: []crud.Field{
			{Key: "name", Label: "Name"},
			{Key: "category", Label: "Category", Kind: crud.Select, Options: Categories},
			{Key: "status", Label: "Status", Kind: crud.Select, Options: []string{"active", "disabled"}},
		},
		Form: []crud.Field{
			{Key: "name", Label: "Name"},
			{Key: "category", Label: "Category", Kind: crud.Select, Options: Categories},
			{Key: "price", Label: "Price", Kind: crud.Number},
			{Key: "active", Label: "Active", Kind: crud.Check},
		},
	})
}
