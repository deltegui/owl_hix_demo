package web

import (
	"strconv"

	"github.com/deltegui/hx"
	"github.com/deltegui/hxbs"
)

func layoutUtilities() hx.INode {
	return bs.Stack(3,
		bs.Flex(bs.Badge("left", bs.Primary), bs.Spacer(), bs.Badge("right", bs.Secondary)).Class(bs.AlignItems(bs.AlignCenter)).Attribute("data-demo", "flex"),
		bs.Center(hx.Span().Text("centered")).Class("border", "p-4"),
		bs.Grid(3, 2, bs.Badge("1", bs.Info), bs.Badge("2", bs.Info), bs.Badge("3", bs.Info)).Attribute("data-demo", "grid"),
		bs.Row(bs.ColOn(bs.MD, 6, hx.Span().Text("a")), bs.ColOn(bs.MD, 6, hx.Span().Text("b"))).Class(bs.Gutter(3)),
		bs.Blockquote("A well-known quote.", "Someone famous"),
		bs.ListInline(hx.Span().Text("one"), hx.Span().Text("two"), hx.Span().Text("three")),
		bs.ListUnstyled(hx.Span().Text("no bullet")))
}

func moreForms() hx.INode {
	pressed := hx.Signal(false)
	size := hx.Signal("m")
	owner := hx.Signal("Ada Lovelace")
	email := hx.Signal("")
	valid := hx.Computed(func() bool { return email.Get() != "" })
	a, b := hx.Signal(true), hx.Signal(false)
	inline, role := hx.Signal(false), hx.Signal("user")
	colour := hx.Signal("#ff8800")
	opted := hx.Signal(false)
	city := hx.Signal("")
	state := hx.Computed(func() string {
		return "pressed=" + strconv.FormatBool(pressed.Get()) + " size=" + size.Get() + " a=" + strconv.FormatBool(a.Get()) +
			" b=" + strconv.FormatBool(b.Get()) + " inline=" + strconv.FormatBool(inline.Get()) + " role=" + role.Get() +
			" colour=" + colour.Get() + " opted=" + strconv.FormatBool(opted.Get())
	})
	return bs.Stack(3,
		bs.FormRow(
			bs.ColOn(bs.MD, 6, hx.Div().Body(bs.Label("x-owner", "Owner"), bs.PlainText("x-owner", owner))),
			bs.ColOn(bs.MD, 6, hx.Div().Body(bs.Label("x-email", "Email"),
				bs.IsValid(bs.Input("email", "x-email", email), valid), bs.ValidFeedback("Looks good!")))),
		hx.Div().Class("d-flex", "gap-3", "flex-wrap").Body(
			bs.ButtonCheck("x-toggle", "Toggle me", pressed),
			bs.ButtonRadios("size", size, []bs.Option{{Value: "s", Label: "S"}, {Value: "m", Label: "M"}, {Value: "l", Label: "L"}}, bs.Outline(bs.Secondary))),
		bs.CheckboxGroup("Extras", bs.CheckItem{ID: "x-a", Label: "Extra A", Value: a}, bs.CheckItem{ID: "x-b", Label: "Extra B", Value: b}),
		hx.Div().Body(bs.CheckInline("x-inline", "Inline check", inline), bs.RadiosInline("role", role, []bs.Option{{Value: "user", Label: "User"}, {Value: "admin", Label: "Admin"}})),
		bs.Fieldset("Colour and city", bs.ColorInput("x-colour", colour),
			bs.Input("text", "x-city", city).Attribute("list", "x-cities"), bs.Datalist("x-cities", "Madrid", "Barcelona", "Sevilla")),
		bs.InputGroupSize(bs.Sm, bs.InputGroupCheck("opt in", opted), bs.Input("text", "x-opt", hx.Signal("")).Placeholder("Small group"), bs.Button("Go", bs.Outline(bs.Secondary))),
		hx.Pre().Attribute("data-demo", "more-state").BindText(state))
}

func moreComponents() hx.INode {
	loading := hx.Signal(false)
	tab := hx.Signal("one")
	seg := hx.Signal(25)
	return bs.Stack(3,
		bs.Alert(hx.Div().Body(bs.AlertHeading("Well done"), hx.P().Class("mb-0").Body(hx.Span().Text("See "), bs.AlertLink("/bs", "this link"))), bs.Success),
		bs.CardGrid(3,
			bs.Card(bs.CardHeaderTabs(tab, []bs.Tab{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}),
				bs.CardBody(bs.TabPanel(tab, "one", func() hx.INode { return hx.P().Text("Card tab one") }),
					bs.TabPanel(tab, "two", func() hx.INode { return hx.P().Text("Card tab two") }))),
			bs.Card(bs.CardBody(bs.CardTitle("Clickable"), bs.CardText("The whole card is a link."), bs.StretchedLink("/products", "Products"))).Class("position-relative"),
			bs.Card(bs.CardBody(bs.CardTitle("Overlay")), bs.CardImageOverlay(hx.P().Text("Over the image")))),
		bs.CardGroup(bs.Card(bs.CardBody(bs.CardText("Group 1"))), bs.Card(bs.CardBody(bs.CardText("Group 2")))),
		hx.Div().Class("d-flex", "gap-4", "align-items-center").Body(
			bs.WithBadge(bs.Button("Inbox", bs.Primary), "99+"),
			bs.LoadingButton("Save", loading, bs.Success).OnClick(func(hx.EventContext) { loading.Set(true) }),
			bs.Button("Done", bs.Outline(bs.Secondary), bs.Sm).OnClick(func(hx.EventContext) { loading.Set(false) })),
		bs.ProgressStacked(bs.ProgressSegmentOf(seg, 100), bs.ProgressSegment(30, 100, bs.Success, bs.Striped), bs.ProgressSegment(20, 100, bs.Info)),
		hx.Div().Class("d-flex", "gap-2").Body(
			bs.Button("seg -5", bs.Outline(bs.Secondary), bs.Sm).OnClick(func(hx.EventContext) { seg.Set(max(0, seg.Get()-5)) }),
			bs.Button("seg +5", bs.Outline(bs.Secondary), bs.Sm).OnClick(func(hx.EventContext) { seg.Set(min(50, seg.Get()+5)) })),
		bs.Table([]bs.Mod{bs.Bordered}, bs.TableCaption("List of users"),
			hx.TBody().Body(
				bs.TableRow([]bs.Mod{bs.Danger}, hx.Td().Text("Danger row")),
				bs.TableRow([]bs.Mod{bs.Success}, hx.Td().Text("Success row")))))
}

func moreMenus() hx.INode {
	split, up, navOpen, navbar := hx.Signal(false), hx.Signal(false), hx.Signal(false), hx.Signal(false)
	end := hx.Signal(false)
	last := hx.Signal("none")
	pick := func(name string, close *hx.SignalT[bool]) hx.INode {
		return bs.DropdownItem(name, "", func() { last.Set(name); close.Set(false) })
	}
	modal, full := hx.Signal(false), hx.Signal(false)
	return bs.Stack(3,
		hx.Div().Class("d-flex", "gap-3", "align-items-center", "flex-wrap").Body(
			bs.SplitDropdown(split, "Action", func() { last.Set("main") }, []hx.INode{pick("split-a", split), pick("split-b", split)}, bs.Danger),
			bs.Dropdown(up, "Dropup", []hx.INode{pick("up-a", up), pick("up-b", up)}, bs.Dropup, bs.Info),
			bs.Dropdown(end, "Menu end", []hx.INode{pick("end-a", end)}, bs.MenuEnd, bs.Outline(bs.Secondary)),
			hx.P().Class("mb-0").Body(hx.Span().Text("Last: "), hx.Strong().Attribute("data-demo", "last").BindText(last))),
		bs.Navbar("lg",
			bs.NavbarBrand("Second", "/bs"),
			bs.NavbarToggler(navbar),
			bs.NavbarCollapse(navbar, bs.NavbarNav(
				bs.NavbarLink("Home", "/", true),
				bs.NavDropdown(navOpen, "More", []hx.INode{pick("nav-a", navOpen), pick("nav-b", navOpen)}),
				bs.NavbarText("A text")))).Class("bg-body-tertiary", "rounded", "px-3").Attribute("data-demo", "second-navbar"),
		hx.Div().Class("d-flex", "gap-2").Body(
			bs.Button("Scrollable modal", bs.Secondary).OnClick(func(hx.EventContext) { modal.Set(true) }),
			bs.Button("Fullscreen modal", bs.Dark).OnClick(func(hx.EventContext) { full.Set(true) })),
		bs.Modal(modal, "Long text", func() { modal.Set(false) }, func() hx.INode {
			return hx.Div().Style("height", "1500px").Text("A very tall body")
		}, nil, bs.Scrollable),
		bs.ModalCentered(full, "Everything", func() { full.Set(false) }, func() hx.INode { return hx.P().Text("Fills the screen.") }, nil, bs.Fullscreen))
}
