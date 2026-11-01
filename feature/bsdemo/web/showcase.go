// Package web is the Bootstrap showcase: one screen with every component of
// hxbs. It compiles on the server (which writes it) and in the browser (which
// hydrates it, so the modals, menus, tabs and toasts work).
package web

import (
	"strconv"

	"github.com/deltegui/hx"
	"github.com/deltegui/hxbs"
)

// Island is the whole showcase. It needs no state.
func Island(string) hx.INode {
	return bs.Container(
		hx.Div().Class("my-4").Body(bs.Heading(1, "hxbs"), bs.Lead("Every Bootstrap component as a Hix node")),
		section("Buttons and badges", buttons()),
		section("Alert, card, list group, progress, spinner", surfaces()),
		section("Form", forms()),
		section("Tabs, accordion, breadcrumb, pagination", navigation()),
		section("Modal, offcanvas, toast", overlays()),
		section("Dropdown, collapse, navbar, carousel", menus()),
		section("Table", table()),
		section("Layout utilities, quote, lists", layoutUtilities()),
		section("More form controls", moreForms()),
		section("More components", moreComponents()),
		section("More menus and modals", moreMenus()),
	)
}

func section(title string, body hx.INode) hx.INode {
	return hx.El("section").Attribute("data-section", title).Class("mb-5").Body(
		bs.Heading(2, title).Class("h4", "border-bottom", "pb-2", "mb-3"), body)
}

func buttons() hx.INode {
	return hx.Div().Class("d-flex", "flex-column", "gap-3").Body(
		hx.Div().Class("d-flex", "gap-2", "flex-wrap").Body(
			bs.Button("Primary"), bs.Button("Secondary", bs.Secondary), bs.Button("Success", bs.Success),
			bs.Button("Danger", bs.Danger), bs.Button("Warning", bs.Warning), bs.Button("Link", bs.Link),
			bs.Button("Outline", bs.Outline(bs.Danger))),
		hx.Div().Class("d-flex", "gap-2", "align-items-center", "flex-wrap").Body(
			bs.Button("Small", bs.Sm), bs.Button("Large", bs.Lg), bs.ButtonLink("/products", "A link", bs.Info),
			bs.ButtonGroup(nil, bs.Button("One"), bs.Button("Two"), bs.Button("Three"))),
		hx.Div().Class("d-flex", "gap-2").Body(
			bs.Badge("Badge", bs.Primary), bs.Badge("Pill", bs.Success, bs.Pill), bs.Badge("Danger", bs.Danger)))
}

func surfaces() hx.INode {
	closed := hx.Signal(false)
	progress := hx.Signal(40)
	return hx.Div().Class("d-flex", "flex-column", "gap-3").Body(
		bs.Alert(hx.P().Class("mb-0").Text("A simple alert"), bs.Primary),
		hx.Dynamic(func() hx.INode {
			if closed.Get() {
				return nil
			}
			return bs.DismissibleAlert(func() { closed.Set(true) }, hx.P().Class("mb-0").Text("A dismissible alert"), bs.Warning)
		}),
		bs.RowCols(2,
			bs.Col(bs.Card(bs.CardHeader(hx.Span().Text("Card header")),
				bs.CardBody(bs.CardTitle("Card title"), bs.CardSubtitle("Card subtitle"), bs.CardText("Some text."), bs.Button("Go")),
				bs.CardFooter(hx.Span().Class("text-body-secondary").Text("Footer")))),
			bs.Col(bs.ListGroup([]bs.Mod{bs.Flush},
				bs.ListGroupItem(hx.Span().Text("First item")),
				bs.ListGroupItem(hx.Span().Text("Second item"), bs.Success),
				bs.ListGroupAction("An action", nil))),
		),
		bs.ProgressOf(progress, 100, bs.Striped),
		hx.Div().Class("d-flex", "gap-2", "align-items-center").Body(
			bs.Button("-10", bs.Outline(bs.Secondary), bs.Sm).OnClick(func(hx.EventContext) { progress.Set(max(0, progress.Get()-10)) }),
			bs.Button("+10", bs.Outline(bs.Secondary), bs.Sm).OnClick(func(hx.EventContext) { progress.Set(min(100, progress.Get()+10)) }),
			bs.Spinner(bs.Primary), bs.Spinner(bs.Grow, bs.Danger, bs.Sm)),
		bs.PlaceholderGlow(bs.Placeholder(6)))
}

func forms() hx.INode {
	name, email, bio, kind := hx.Signal(""), hx.Signal(""), hx.Signal(""), hx.Signal("")
	agree, notify := hx.Signal(false), hx.Signal(true)
	plan, volume := hx.Signal("free"), hx.Signal("40")
	nameErrors := hx.Signal([]string{"required"})
	hx.EffectFunc(func() {
		if name.Get() == "" {
			nameErrors.Set([]string{"required"})
		} else {
			nameErrors.Set(nil)
		}
	})
	summary := hx.Computed(func() string {
		return "name=" + name.Get() + " kind=" + kind.Get() + " agree=" + strconv.FormatBool(agree.Get()) +
			" notify=" + strconv.FormatBool(notify.Get()) + " plan=" + plan.Get() + " volume=" + volume.Get()
	})
	return bs.Row(
		bs.ColOn(bs.MD, 6, hx.Form().Body(
			bs.FormField("Name", "f-name", bs.Input("text", "f-name", name), nameErrors),
			bs.FormField("Email", "f-email", bs.Input("email", "f-email", email), nil),
			bs.FormField("Kind", "f-kind", bs.Select("f-kind", kind, "Choose…", bs.Plain("Person", "Company")), nil),
			bs.FormField("Bio", "f-bio", bs.TextArea("f-bio", bio), nil),
			hx.Div().Class("mb-3").Body(
				bs.Checkbox("f-agree", "I agree", agree),
				bs.Switch("f-notify", "Notifications", notify)),
			hx.Div().Class("mb-3").Body(bs.Radios("plan", plan, []bs.Option{{Value: "free", Label: "Free"}, {Value: "pro", Label: "Pro"}})),
			hx.Div().Class("mb-3").Body(bs.Label("f-vol", "Volume"), bs.Range("f-vol", volume)),
			hx.Div().Class("mb-3").Body(bs.InputGroup(bs.InputGroupText(hx.Span().Text("@")), bs.Input("text", "f-user", hx.Signal("")).Placeholder("Username"))),
			hx.Div().Class("mb-3").Body(bs.FloatingLabel("Floating", "f-float", bs.Input("text", "f-float", hx.Signal("")))),
			hx.Div().Class("mb-3").Body(bs.FileInput("f-file", nil)),
			bs.SubmitButton("Submit"))),
		bs.ColOn(bs.MD, 6, bs.Card(bs.CardHeader(hx.Span().Text("Live values")),
			bs.CardBody(hx.Pre().Class("mb-0").BindText(summary)))))
}

func navigation() hx.INode {
	tab, pill := hx.Signal("account"), hx.Signal("one")
	faq := hx.Signal("a")
	page, pages := hx.Signal(3), hx.Signal(12)
	return hx.Div().Class("d-flex", "flex-column", "gap-3").Body(
		bs.Breadcrumb(bs.Crumb{Label: "Home", Href: "/"}, bs.Crumb{Label: "Components", Href: "/bs"}, bs.Crumb{Label: "Breadcrumb", Current: true}),
		hx.Div().Body(
			bs.Tabs(tab, []bs.Tab{{ID: "account", Label: "Account"}, {ID: "password", Label: "Password"}}),
			hx.Div().Class("p-3").Body(
				bs.TabPanel(tab, "account", func() hx.INode { return hx.P().Text("Account settings.") }),
				bs.TabPanel(tab, "password", func() hx.INode { return hx.P().Text("Change your password.") }))),
		bs.Tabs(pill, []bs.Tab{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}, bs.Pills),
		bs.Accordion(faq,
			bs.AccordionItem{ID: "a", Title: "Is it accessible?", Body: func() hx.INode { return hx.P().Text("Yes.") }},
			bs.AccordionItem{ID: "b", Title: "Does it need JavaScript?", Body: func() hx.INode { return hx.P().Text("No, only Hix signals.") }}),
		bs.Pagination(page, pages, page.Set))
}

func overlays() hx.INode {
	modal, centered, side := hx.Signal(false), hx.Signal(false), hx.Signal(false)
	toast := hx.Signal(false)
	return hx.Div().Class("d-flex", "gap-2").Body(
		bs.Button("Open modal").OnClick(func(hx.EventContext) { modal.Set(true) }),
		bs.Button("Open centered", bs.Secondary).OnClick(func(hx.EventContext) { centered.Set(true) }),
		bs.Button("Open offcanvas", bs.Info).OnClick(func(hx.EventContext) { side.Set(true) }),
		bs.Button("Show toast", bs.Success).OnClick(func(hx.EventContext) { toast.Set(true) }),
		bs.Modal(modal, "Edit profile", func() { modal.Set(false) },
			func() hx.INode { return hx.P().Text("A modal is in the tree only while it is open.") },
			func() hx.INode { return bs.Button("Save").OnClick(func(hx.EventContext) { modal.Set(false) }) }),
		bs.ModalCentered(centered, "Centered", func() { centered.Set(false) },
			func() hx.INode { return hx.P().Text("In the middle.") }, nil, bs.Sm),
		bs.Offcanvas(side, "start", "Menu", func() { side.Set(false) },
			func() hx.INode { return hx.P().Text("An offcanvas from the left.") }),
		bs.ToastContainer("bottom-0 end-0",
			bs.Toast(toast, "Saved", func() { toast.Set(false) }, func() hx.INode { return hx.Span().Text("Event created") })))
}

func menus() hx.INode {
	menu, open, nav := hx.Signal(false), hx.Signal(false), hx.Signal(false)
	chosen := hx.Signal("none")
	slide := hx.Signal(0)
	return hx.Div().Class("d-flex", "flex-column", "gap-3").Body(
		hx.Div().Class("d-flex", "align-items-center", "gap-3").Body(
			bs.Dropdown(menu, "Open menu", []hx.INode{
				bs.DropdownHeader("My account"),
				bs.DropdownItem("Profile", "", func() { chosen.Set("profile"); menu.Set(false) }),
				bs.DropdownItem("Settings", "", func() { chosen.Set("settings"); menu.Set(false) }),
				bs.DropdownDivider(),
				bs.DropdownItem("Docs", "/about", nil)}),
			hx.P().Class("mb-0").Body(hx.Span().Text("Chose: "), hx.Strong().BindText(chosen))),
		hx.Div().Body(
			bs.Button("Toggle collapse", bs.Outline(bs.Primary)).OnClick(func(hx.EventContext) { open.Set(!open.Get()) }),
			bs.Collapse(open, bs.Card(bs.CardBody(hx.Span().Text("Collapsible content"))).Class("mt-2"))),
		bs.Navbar("lg", bs.Container(
			bs.NavbarBrand("Brand", "/bs"),
			bs.NavbarToggler(nav),
			bs.NavbarCollapse(nav, bs.NavbarNav(bs.NavbarLink("Home", "/", true), bs.NavbarLink("Products", "/products", false))))).Class("bg-body-tertiary", "rounded"),
		bs.Carousel(slide,
			bs.CarouselItem(hx.Div().Class("bg-primary", "text-white", "p-5").Text("Slide one"), bs.CarouselCaption("One", "First")),
			bs.CarouselItem(hx.Div().Class("bg-success", "text-white", "p-5").Text("Slide two")),
			bs.CarouselItem(hx.Div().Class("bg-danger", "text-white", "p-5").Text("Slide three"))))
}

func table() hx.INode {
	return bs.TableResponsive(bs.SimpleTable([]bs.Mod{bs.Striped, bs.Hover}, []string{"Invoice", "Status", "Amount"},
		row("INV001", "Paid", "$250.00"), row("INV002", "Pending", "$150.00"), row("INV003", "Unpaid", "$350.00")))
}

func row(cells ...string) hx.INode {
	nodes := make([]hx.INode, len(cells))
	for i, c := range cells {
		nodes[i] = hx.Td().Text(c)
	}
	return hx.Tr().BodyList(nodes)
}
