// Package layout is the page around the features: the side menu, the theme
// selector, the language buttons, and the islands with their state. It is what
// views/shared/layout.html is in LittleCash, written with Hix and rendered by
// OWL's HixRenderer.
package layout

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/hxui"

	"owldemo/web/locale"
	"owldemo/web/ui/bs"
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

// Render writes the page with the renderer of the Mux.
func Render(ctx owl.Ctx, p Page) error {
	return ctx.View(owl.HixView(func(ctx owl.Ctx) hx.INode { return build(ctx, p) }))
}

func build(ctx owl.Ctx, p Page) hx.INode {
	shared := hxui.NewView(ctx, locale.Shared)
	csrfToken, _ := ctx.Get(csrfKey).(string)

	var footer []hx.INode
	islands := p.Islands
	if p.Bundle != "" {
		islands = append(islands[:len(islands):len(islands)], Island{Name: "theme", State: shared.T("Shared.Theme"), Build: themeBuild})
	}
	for _, i := range islands {
		if i.Name == "theme" {
			footer = append(footer, hx.Div().Class("mb-3").Body(host(i)))
		}
	}
	footer = append(footer, languages(ctx, shared))

	var content []hx.INode
	if p.Content != nil {
		content = append(content, p.Content)
	}
	for _, i := range p.Islands {
		content = append(content, host(i))
	}

	body := []hx.INode{
		hx.Div().Class("d-flex", "min-vh-100").Body(
			sidebar(ctx, shared, footer),
			hx.Main().Class("flex-grow-1", "p-4").BodyList(content)),
	}
	for _, i := range islands {
		if i.State != "" {
			body = append(body, hx.Raw(hx.StateScript(bs.StateID(i.Name), i.State)))
		}
	}
	if p.Bundle != "" {
		if p.Chart {
			body = append(body, hx.El("script").Src("https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"))
		}
		body = append(body, hx.El("script").Src("/js/wasm_exec.js"), hx.Raw("<script>"+loader(p.Bundle)+"</script>"))
	}

	return hx.El("html").Attribute("lang", ctx.GetCurrentLanguage()).Body(
		hx.El("head").Body(
			hx.El("meta").Attribute("charset", "utf-8"),
			hx.El("meta").Attribute("name", "viewport").Attribute("content", "width=device-width, initial-scale=1"),
			hx.El("meta").Attribute("name", "csrf-token").Attribute("content", csrfToken),
			hx.El("title").Text(p.Title+" · OWL demo"),
			hx.El("link").Attribute("rel", "stylesheet").Href(bootstrapCSS)),
		hx.El("body").BodyList(body))
}

// host is the element of an island, with the island built inside it.
func host(i Island) hx.INode { return bs.Island(i.Name).Body(i.Build(i.State)) }

func loader(bundle string) string {
	return `const go = new Go();
WebAssembly.instantiateStreaming(fetch("/js/` + bundle + `.wasm"), go.importObject).then((r) => go.run(r.instance));`
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
	for _, lang := range []string{"es", "en"} {
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
