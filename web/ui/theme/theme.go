// Package theme is the selector of Bootstrap's light and dark themes. It is an
// island of its own, in a package of its own: it imports nothing heavy, so a
// binary that only has this island does not carry the JSON code of the screens.
package theme

import (
	"github.com/deltegui/hx"

	"owldemo/web/ui/bs"
)

// Build is Island for a registry of islands. The state is the text of the label,
// which the server translates ("Theme" if it is empty).
func Build(label string) hx.INode { return Island(label) }

// Island is a small island: a selector in the menu that switches Bootstrap's
// light and dark themes and remembers the choice. It knows nothing about the
// screen that is next to it.
//
// The server and the browser both start from "auto", so the HTML matches. Once
// the island is hydrated the browser reads the saved choice.
func Island(label string) hx.INode {
	if label == "" {
		label = "Theme"
	}
	theme := hx.Signal("auto")
	mounted := false
	hx.OnMount(func() {
		theme.Set(saved())
		mounted = true
	})
	box := hx.Div()
	box.Effect(func() {
		choice := theme.Get()
		apply(choice)
		if mounted { // not for the first value, which is not the user's
			store(choice)
		}
	})
	return box.Body(
		hx.Label().Class("form-label", "small", "text-white-50").Text(label),
		bs.Select(theme, "", bs.Plain("auto", "light", "dark")).Class("form-select-sm"))
}
