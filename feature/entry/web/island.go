package entryweb

import (
	"encoding/json"

	"github.com/deltegui/hx"
)

// Island builds the screen from the state the server sent (JSON).
func Island(state string) hx.INode {
	var s State
	if state != "" {
		json.Unmarshal([]byte(state), &s)
	}
	return Build(s)
}

// Build is the whole maintenance: the list and, over it, the editor or the
// viewer of an entry. The list stays alive while the second layer is open, so
// the filter, the order and the page are where the user left them.
func Build(s State) hx.INode {
	a := newApp(s.Texts, s.Lookups)

	main := hx.Div().Attribute("data-layer", "list").Body(list(a, s.Listing))
	main.Effect(func() { main.ClassIf("d-none", a.open.Get().Kind != "") })

	second := hx.Div().Body(hx.Dynamic(func() hx.INode {
		o, entry := a.open.Get(), a.loaded.Get()
		switch {
		case o.Kind == "":
			return nil
		case entry == nil:
			return hx.P().Class("text-muted").Text(a.t.Get("Shared.Loading"))
		}
		var node hx.INode
		hx.Untrack(func() {
			if o.Kind == "view" {
				node = viewer(a, *entry)
			} else {
				node = editor(a, *entry)
			}
		})
		return node
	}))

	return hx.Div().Body(main, second)
}
