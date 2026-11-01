// Package layout is the page around the features: the side menu, the theme
// selector, the language buttons, and the islands with their state. It is what
// views/shared/layout.html is in LittleCash, written with Hix and rendered by
// ctx.View.
package layout

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"
	"github.com/deltegui/owl/v3/island"

	"owldemo/web/locale"
)

// MenuItem is a link of the side menu. The label is already translated.
type MenuItem struct{ Href, Label string }

// Menu gives the items of the side menu. The menu feature sets it when it
// starts, like LittleCash's menu sets TemplateFuncs["GetMenu"].
var Menu func(ctx owl.Ctx) []MenuItem

// Island is a piece of the page that the browser hydrates. State is the text
// the server built it from; the browser builds the same thing from it.
type Island struct {
	Name  string
	State string
	Build func(state string) hx.INode
}

// Of is an island defined with island.Define, placed with its state:
// layout.Of(island.Mount(productweb.List, state)).
func Of(i island.Instance) Island { return Island{Name: i.Name(), State: i.State, Build: i.Build} }

// Page is what a handler gives to Render.
type Page struct {
	Title string // translated
	// Bundle is the WASM the page loads: "lite", "full" or "forms" (see client/).
	// Empty means the page loads none.
	Bundle  string
	Chart   bool // loads Chart.js
	Islands []Island
	Content hx.INode // HTML that is not an island, shown instead of them
}

const bootstrapCSS = "https://cdn.jsdelivr.net/npm/bootstrap@5.3.8/dist/css/bootstrap.min.css"

// Render writes the page with ctx.View.
func Render(ctx owl.Ctx, p Page) error {
	return ctx.View(owl.HixView(func(ctx owl.Ctx) hx.INode { return build(ctx, p) }))
}

func build(ctx owl.Ctx, p Page) hx.INode {
	shared := hxui.NewView(ctx, locale.Shared)

	// The theme selector goes in the menu of every page that has a bundle.
	var footer []hx.INode
	var instances []island.Instance
	for _, i := range p.Islands {
		instances = append(instances, i.instance())
	}
	if p.Bundle != "" {
		theme := Island{Name: "theme", State: shared.T("Shared.Theme"), Build: themeBuild}.instance()
		instances = append(instances, theme)
		footer = append(footer, hx.Div().Class("mb-3").Body(theme.Node()))
	}
	footer = append(footer, languages(ctx, shared))

	var content []hx.INode
	if p.Content != nil {
		content = append(content, p.Content)
	}
	for _, i := range p.Islands {
		content = append(content, host(i))
	}

	opts := hxui.AppOptions{
		Title:   p.Title + " · OWL demo",
		Head:    []hx.INode{hxui.Stylesheet(bootstrapCSS)},
		Bundle:  p.Bundle,
		Islands: instances,
	}
	if p.Chart {
		opts.Scripts = []hx.INode{hx.El("script").Src("https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js")}
	}
	return hxui.App(ctx, opts,
		hx.Div().Class("d-flex", "min-vh-100").Body(
			sidebar(ctx, shared, footer),
			hx.Main().Class("flex-grow-1", "p-4").BodyList(content)))
}

// host is the element of an island, with the island built inside it.
func host(i Island) hx.INode { return i.instance().Node() }

func (i Island) instance() island.Instance {
	return island.Instance{Island: island.Func(i.Name, i.Build), State: i.State}
}

func sidebar(ctx owl.Ctx, shared hxui.View, footer []hx.INode) hx.INode {
	var items []MenuItem
	if Menu != nil {
		items = Menu(ctx)
	}
	links := make([]hx.INode, 0, len(items))
	for _, item := range items {
		links = append(links, hx.Li().Class("nav-item").Body(
			hx.A().Href(item.Href).Class("nav-link", "text-white").
				ClassIf("active", item.Href == ctx.Req.URL.Path).Text(item.Label)))
	}
	parts := []hx.INode{
		hx.A().Href("/").Class("fs-4", "mb-3", "text-white", "text-decoration-none").Text(shared.T("Shared.Brand")),
		hx.Hr(),
		hx.Ul().Class("nav", "nav-pills", "flex-column").BodyList(links),
		hx.Hr(),
	}
	return hx.Nav().Class("d-flex", "flex-column", "flex-shrink-0", "p-3", "bg-dark", "text-white").
		Style("width", "220px").BodyList(append(parts, footer...))
}

// languages are the buttons that change the language of the cookie. They are
// forms, so they work without JavaScript, and they carry the CSRF token.
func languages(ctx owl.Ctx, shared hxui.View) hx.INode {
	buttons := []hx.INode{}
	for _, lang := range locale.Languages {
		buttons = append(buttons, hx.Form().Attribute("method", "post").Attribute("action", "/language/"+lang).Class("d-inline").Body(
			shared.CsrfInput(),
			hx.Input().Type("hidden").Attribute("name", "Next").Value(ctx.Req.URL.RequestURI()),
			hx.Button().Type("submit").
				Class("btn", "btn-sm", "me-1").
				ClassIf("btn-light", ctx.GetCurrentLanguage() == lang).
				ClassIf("btn-outline-light", ctx.GetCurrentLanguage() != lang).
				Text(shared.T("Shared."+lang))))
	}
	return hx.Div().Body(hx.Label().Class("form-label", "small", "text-white-50").Text(shared.T("Shared.Language")), hx.Div().BodyList(buttons))
}
