// Package web is the Bulma showcase: one screen with every component of
// hxbulma. It compiles on the server (which writes it) and in the browser
// (which hydrates it, so the modal, the dropdown, the tabs and the burger work).
package web

import (
	"strconv"

	"github.com/deltegui/hx"
	"github.com/deltegui/hxbulma"
)

// Island is the whole showcase. It needs no state.
func Island(string) hx.INode {
	burger := hx.Signal(false)
	return hx.Div().Body(
		topbar(burger),
		bulma.SimpleHero("hxbulma", "Every Bulma component as a Hix node").Class(bulma.Primary),
		bulma.Section(bulma.Container(
			section("Elements", elements()),
			section("Form", formDemo()),
			section("Components", components()),
			section("Layout", layout()),
			section("Helpers", helpers()),
		)),
		bulma.Footer(bulma.Content(hx.P().Class(bulma.TextCentered).Body(
			hx.Strong().Text("hxbulma"), hx.Span().Text(" by Hix. CSS by Bulma.")))))
}

func section(title string, body hx.INode) hx.INode {
	return bulma.Block(bulma.Title(3, title), body, hx.Hr())
}

func topbar(open *hx.SignalT[bool]) hx.INode {
	return bulma.Navbar(
		bulma.NavbarBrand(bulma.NavbarItem("hxbulma", "/bulma"), bulma.NavbarBurger(open)),
		bulma.NavbarMenu(open,
			bulma.NavbarStart(
				bulma.NavbarItem("Elements", "#elements"),
				bulma.NavbarDropdown("OWL demo",
					bulma.NavbarItem("Products", "/products"),
					bulma.NavbarItem("Contacts", "/contacts"),
					bulma.NavbarDivider(),
					bulma.NavbarItem("About", "/about"))),
			bulma.NavbarEnd(bulma.NavbarBox(bulma.Buttons(bulma.ButtonLink("/products", "Back to the demo").Class(bulma.Light))))))
}

func elements() hx.INode {
	notes := hx.Signal([]string{"Saved", "Careful", "Failed"})
	dismiss := func(i int) {
		next := append([]string(nil), notes.Get()[:i]...)
		notes.Set(append(next, notes.Get()[i+1:]...))
	}
	loading := hx.Signal(false)
	progress := hx.Signal(40)
	tags := hx.Signal([]string{"go", "wasm", "bulma"})

	return hx.Div().Id("elements").Body(
		bulma.Title(5, "Buttons"),
		bulma.Buttons(
			bulma.Button("Primary").Class(bulma.Primary),
			bulma.Button("Link").Class(bulma.Link),
			bulma.Button("Info").Class(bulma.Info),
			bulma.Button("Success").Class(bulma.Success),
			bulma.Button("Warning").Class(bulma.Warning),
			bulma.Button("Danger").Class(bulma.Danger),
			bulma.Button("Outlined").Class(bulma.Primary, bulma.Outlined),
			bulma.Button("Rounded").Class(bulma.Info, bulma.Rounded),
			bulma.Button("Small").Class(bulma.Small),
			bulma.Button("Large").Class(bulma.Large)),
		bulma.Buttons(
			bulma.BindClass(bulma.Button("Click: loading").Class(bulma.Primary).OnClick(func(hx.EventContext) { loading.Set(!loading.Get()) }), bulma.Loading, loading),
			bulma.Buttons(bulma.Button("One"), bulma.Button("Two"), bulma.Button("Three")).Class(bulma.HasAddons)),
		bulma.Title(5, "Notifications and tags"),
		hx.Dynamic(func() hx.INode {
			var items []hx.INode
			colors := []string{bulma.Success, bulma.Warning, bulma.Danger}
			for i, text := range notes.Get() {
				items = append(items, bulma.DismissibleNotification(func() { dismiss(i) }, hx.Span().Text(text)).Class(colors[i%3], bulma.Light))
			}
			return hx.Noop().BodyList(items)
		}),
		hx.Dynamic(func() hx.INode {
			var items []hx.INode
			for i, tag := range tags.Get() {
				items = append(items, bulma.DeletableTag(tag, func() {
					current := tags.Get()
					tags.Set(append(append([]string(nil), current[:i]...), current[i+1:]...))
				}).Class(bulma.Info, bulma.Light))
			}
			return bulma.Tags().BodyList(items)
		}),
		bulma.Tags(bulma.Tag("Version").Class(bulma.Dark), bulma.Tag("1.0.4").Class(bulma.Info)).Class(bulma.HasAddons),
		bulma.Title(5, "Progress"),
		bulma.ProgressOf(progress, 100).Class(bulma.Primary),
		bulma.Buttons(
			bulma.Button("-10").Class(bulma.Small).OnClick(func(hx.EventContext) { progress.Set(max(0, progress.Get()-10)) }),
			bulma.Button("+10").Class(bulma.Small).OnClick(func(hx.EventContext) { progress.Set(min(100, progress.Get()+10)) })),
		bulma.Indeterminate().Class(bulma.Info, bulma.Small),
		bulma.Title(5, "Box, content, icon, skeleton"),
		bulma.Box(bulma.Content(hx.H1().Text("Plain HTML"), hx.P().Text("Headings, "), hx.Ul().Body(hx.Li().Text("lists"), hx.Li().Text("and quotes are styled by Content.")))),
		bulma.Buttons(bulma.IconText("★", "IconText")),
		bulma.Columns(
			bulma.Column(bulma.SkeletonBlock()),
			bulma.Column(bulma.SkeletonLines(4))),
		bulma.Title(5, "Image"),
		bulma.Columns(bulma.Column(bulma.Image("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='256' height='256'><rect width='256' height='256' fill='%2300d1b2'/></svg>", "Square").Class("is-128x128"))),
		bulma.Title(5, "Table"),
		bulma.TableContainer(bulma.SimpleTable([]string{"Component", "Tag"},
			row("Button", "button"), row("Box", "div"), row("Progress", "progress")).Class(bulma.Striped, bulma.Hoverable, bulma.Fullwidth)))
}

func row(cells ...string) hx.INode {
	tr := hx.Tr()
	nodes := make([]hx.INode, len(cells))
	for i, c := range cells {
		nodes[i] = hx.Td().Text(c)
	}
	return tr.BodyList(nodes)
}

func formDemo() hx.INode {
	name, email, bio := hx.Signal(""), hx.Signal(""), hx.Signal("")
	kind := hx.Signal("")
	agree := hx.Signal(false)
	plan := hx.Signal("free")
	chosen := hx.Signal("")
	errs := hx.Computed(func() string {
		if name.Get() == "" {
			return "required"
		}
		return ""
	})
	nameErrors := hx.Signal([]string{})
	hx.EffectFunc(func() {
		if errs.Get() != "" {
			nameErrors.Set([]string{errs.Get()})
		} else {
			nameErrors.Set(nil)
		}
	})
	summary := hx.Computed(func() string {
		return "name=" + name.Get() + " email=" + email.Get() + " kind=" + kind.Get() + " agree=" + strconv.FormatBool(agree.Get()) + " plan=" + plan.Get()
	})

	return bulma.Columns(
		bulma.Column(
			bulma.FormField("Name", bulma.Input("text", name).Class(bulma.Primary), nameErrors),
			bulma.Field(bulma.Label("Email"), bulma.Control(
				bulma.Input("email", email), hx.Span().Class("icon", bulma.Small, bulma.Left).Text("@")).Class(bulma.HasIconsLeft)),
			bulma.Field(bulma.Label("Kind"), bulma.Control(bulma.Select(kind, "Choose…", bulma.Plain("Person", "Company")))),
			bulma.Field(bulma.Label("Bio"), bulma.Control(bulma.TextArea(bio).Attribute("rows", "2"))),
			bulma.Field(bulma.Control(bulma.Checkbox("I agree", agree))),
			bulma.Field(bulma.Label("Plan"), bulma.Radios("plan", plan, []bulma.Option{{Value: "free", Label: "Free"}, {Value: "pro", Label: "Pro"}})),
			bulma.Field(bulma.Label("File"), bulma.File("Choose a file…", chosen, func(hx.EventContext) { chosen.Set("file chosen") })),
			bulma.Addons(
				bulma.Control(bulma.Input("search", hx.Signal("")).Placeholder("Search")).Class("is-expanded"),
				bulma.Control(bulma.Button("Go").Class(bulma.Info))),
			bulma.Grouped(bulma.Control(bulma.SubmitButton("Submit").Class(bulma.Link)), bulma.Control(bulma.Button("Cancel").Class(bulma.Light)))),
		bulma.Column(bulma.Notification(hx.Pre().BindText(summary)).Class(bulma.Info, bulma.Light),
			bulma.HorizontalField("Horizontal", bulma.Field(bulma.Control(bulma.Input("text", hx.Signal(""))))),
		))
}

func components() hx.INode {
	tab := hx.Signal("overview")
	open := hx.Signal(false)
	card := hx.Signal(false)
	drop := hx.Signal(false)
	page := hx.Signal(3)
	pages := hx.Signal(12)
	card2 := hx.Signal(true)
	message := hx.Signal(true)

	return hx.Div().Body(
		bulma.Title(5, "Breadcrumb"),
		bulma.Breadcrumb(bulma.Crumb{Label: "Home", Href: "/"}, bulma.Crumb{Label: "Components", Href: "/bulma"}, bulma.Crumb{Label: "Breadcrumb", Href: "#", Current: true}).Class(bulma.HasArrowSeparator),
		bulma.Title(5, "Tabs"),
		bulma.Tabs(tab, bulma.Tab{ID: "overview", Label: "Overview"}, bulma.Tab{ID: "usage", Label: "Usage"}, bulma.Tab{ID: "api", Label: "API"}).Class(bulma.Boxed),
		bulma.TabPanel(tab, "overview", func() hx.INode {
			return hx.P().Text("Tabs are a signal: the panels are in the tree only while they are chosen.")
		}),
		bulma.TabPanel(tab, "usage", func() hx.INode {
			return hx.P().Text("bulma.Tabs(active, tabs...) and bulma.TabPanel(active, id, build).")
		}),
		bulma.TabPanel(tab, "api", func() hx.INode { return hx.P().Text("No JavaScript from Bulma is loaded.") }),
		bulma.Title(5, "Dropdown and modal"),
		bulma.Buttons(
			bulma.Dropdown(drop, "Dropdown",
				bulma.DropdownItem("Open the modal", "", func() { drop.Set(false); open.Set(true) }),
				bulma.DropdownItem("Open the card modal", "", func() { drop.Set(false); card.Set(true) }),
				bulma.DropdownDivider(),
				bulma.DropdownItem("A link", "/about", nil)),
			bulma.Button("Modal").Class(bulma.Primary).OnClick(func(hx.EventContext) { open.Set(true) })),
		bulma.Modal(open, func() { open.Set(false) }, func() hx.INode {
			return bulma.Box(hx.P().Text("A modal is in the tree only while it is open."))
		}),
		bulma.ModalCard(card, "Delete it?", func() { card.Set(false) },
			func() hx.INode { return hx.P().Text("This cannot be undone.") },
			func() hx.INode {
				return bulma.Buttons(
					bulma.Button("Delete").Class(bulma.Danger).OnClick(func(hx.EventContext) { card.Set(false) }),
					bulma.Button("Cancel").OnClick(func(hx.EventContext) { card.Set(false) }))
			}),
		bulma.Title(5, "Message"),
		hx.Dynamic(func() hx.INode {
			if !message.Get() {
				return nil
			}
			return bulma.Message("Heads up", func() { message.Set(false) }, hx.P().Text("This message can be dismissed.")).Class(bulma.Warning)
		}),
		bulma.Title(5, "Card"),
		hx.Dynamic(func() hx.INode {
			var body hx.INode
			if card2.Get() {
				body = bulma.CardContent(hx.P().Text("The card header arrow folds this."))
			}
			return bulma.Card(
				bulma.CardHeader("Card", bulma.CardHeaderIcon("▾", func() { card2.Set(!card2.Get()) })),
				body,
				bulma.CardFooter(bulma.CardFooterItem("Save", "", func() {}), bulma.CardFooterItem("Docs", "https://bulma.io", nil)))
		}),
		bulma.Title(5, "Pagination"),
		bulma.Pagination(page, pages, page.Set),
		bulma.Title(5, "Menu and panel"),
		bulma.Columns(
			bulma.Column(bulma.Menu(
				bulma.MenuLabel("General"),
				bulma.MenuList(bulma.MenuLink("Dashboard", "#", true), bulma.MenuLink("Customers", "#", false)),
				bulma.MenuLabel("Admin"),
				bulma.MenuList(bulma.MenuLink("Team", "#", false, bulma.MenuList(bulma.MenuLink("Members", "#", false)))))),
			bulma.Column(bulma.Panel("Repositories",
				bulma.PanelTabs(hx.A().Class(bulma.Active).Text("all"), hx.A().Text("public"), hx.A().Text("private")),
				bulma.PanelLink("", "hx", "#"),
				bulma.PanelLink("", "bulma", "#").Class(bulma.Active),
				bulma.PanelBlock(bulma.Button("Reset").Class(bulma.Link, bulma.Outlined, bulma.Fullwidth))))))
}

func layout() hx.INode {
	return hx.Div().Body(
		bulma.Title(5, "Columns"),
		bulma.Columns(
			bulma.Column(bulma.Notification(hx.Span().Text("One quarter")).Class(bulma.Primary)).Class(bulma.OneQuarter),
			bulma.Column(bulma.Notification(hx.Span().Text("Fills the rest")).Class(bulma.Info))),
		bulma.Title(5, "Level"),
		bulma.Level(
			bulma.LevelLeft(bulma.LevelItem(hx.P().Text("3 items"))),
			bulma.LevelRight(bulma.LevelItem(bulma.Button("Add").Class(bulma.Success)))),
		bulma.Title(5, "Media"),
		bulma.Box(bulma.Media(bulma.Image("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='64' height='64'><circle cx='32' cy='32' r='32' fill='%23485fc7'/></svg>", "avatar").Class("is-64x64"),
			hx.P().Body(hx.Strong().Text("Ada"), hx.Br(), hx.Span().Text("A comment with an image on the side.")), nil)),
		bulma.Title(5, "Tile"),
		bulma.Tile(
			bulma.Tile(bulma.Notification(hx.Span().Text("A")).Class(bulma.Primary)).Class(bulma.Parent),
			bulma.Tile(bulma.Notification(hx.Span().Text("B")).Class(bulma.Warning)).Class(bulma.Parent),
			bulma.Tile(bulma.Notification(hx.Span().Text("C")).Class(bulma.Danger)).Class(bulma.Parent)).Class(bulma.Ancestor),
		bulma.Title(5, "Grid"),
		bulma.FixedGrid(bulma.Grid(
			bulma.Cell(bulma.Notification(hx.Span().Text("1"))), bulma.Cell(bulma.Notification(hx.Span().Text("2"))), bulma.Cell(bulma.Notification(hx.Span().Text("3")))),
		).Class(bulma.HasCols(3)))
}

func helpers() hx.INode {
	return bulma.Box(
		hx.P().Class(bulma.TextColor(bulma.Primary), bulma.TextWeightBold).Text("has-text-primary, has-text-weight-bold"),
		hx.P().Class(bulma.BackgroundColor(bulma.Warning), bulma.Spacing("p", "", 2)).Text("has-background-warning, p-2"),
		hx.P().Class(bulma.Uppercase, bulma.TextCentered).Text("uppercase and centered"))
}
