// Package web is the Basecoat showcase: one screen with every component of
// hxcoat. It compiles on the server (which writes it) and in the browser (which
// hydrates it, so the dialogs, menus, tabs and toasts work).
package web

import (
	"strconv"

	"github.com/deltegui/hx"
	"github.com/deltegui/hxcoat"
)

// Island is the whole showcase. It needs no state.
func Island(string) hx.INode {
	return hx.Div().Class("mx-auto", "max-w-4xl", "p-6", "flex", "flex-col", "gap-10").Body(
		hx.Div().Body(
			hx.H1().Class("text-3xl", "font-bold").Text("hxcoat"),
			hx.P().Class("text-muted-foreground").Text("Every Basecoat component as a Hix node")),
		section("Buttons and badges", buttons()),
		section("Alert, card, avatar, item, empty", surfaces()),
		section("Form", forms()),
		section("Tabs, accordion, breadcrumb, pagination", navigation()),
		section("Dialog, alert dialog, drawer", dialogs()),
		section("Dropdown, popover, select, combobox, command, tooltip", menus()),
		section("Table, progress, skeleton, toast", data()),
	)
}

func section(title string, body hx.INode) hx.INode {
	return hx.El("section").Attribute("data-section", title).Class("flex", "flex-col", "gap-4").Body(
		hx.H2().Class("text-xl", "font-semibold", "border-b", "pb-2").Text(title), body)
}

func buttons() hx.INode {
	return coat.Stack(4,
		coat.Row(2,
			coat.Button("Primary"), coat.Button("Secondary", coat.Secondary), coat.Button("Outline", coat.Outline),
			coat.Button("Ghost", coat.Ghost), coat.Button("Link", coat.Link), coat.Button("Destructive", coat.Destructive)),
		coat.Row(2, coat.Button("Small", coat.Sm), coat.Button("Large", coat.Lg), coat.ButtonLink("/products", "A link", coat.Outline),
			coat.ButtonGroup(coat.Button("One", coat.Outline), coat.Button("Two", coat.Outline), coat.Button("Three", coat.Outline))),
		coat.Row(2, coat.Badge("Badge"), coat.Badge("Secondary", coat.Secondary), coat.Badge("Outline", coat.Outline), coat.Badge("Destructive", coat.Destructive), coat.Kbd("⌘K")))
}

func surfaces() hx.INode {
	return coat.Stack(4,
		coat.Alert("Heads up", "You can add components with Hix.").Class("max-w-xl"),
		coat.Alert("Error", "Your session has expired.", coat.Destructive).Class("max-w-xl"),
		coat.Grid(2, 4,
			coat.Card(
				coat.CardHeader("Card title", "Card description", coat.Button("Action", coat.Outline, coat.Sm)),
				coat.CardSection(hx.P().Text("Card content")),
				coat.CardFooter(coat.Button("Save"), coat.Button("Cancel", coat.Outline))),
			coat.Stack(3,
				coat.Row(2, coat.Avatar("", "", "AB"), coat.Avatar("", "", "CD", coat.Lg),
					coat.AvatarGroup(coat.Avatar("", "", "E"), coat.Avatar("", "", "F"), coat.Avatar("", "", "G"))),
				coat.ItemGroup(
					coat.Item(coat.ItemOptions{Title: "Ada Lovelace", Description: "Mathematician", Media: coat.Avatar("", "", "AL"),
						Actions: coat.Button("Follow", coat.Outline, coat.Sm), Mods: []coat.Mod{coat.Outline}}),
					coat.Item(coat.ItemOptions{Title: "Alan Turing", Description: "Computer scientist", Mods: []coat.Mod{coat.Muted}})))),
		coat.Empty("No projects yet", "Create your first project to get started.", coat.Button("New project")))
}

func forms() hx.INode {
	name, email, bio := hx.Signal(""), hx.Signal(""), hx.Signal("")
	kind := hx.Signal("")
	agree, notify := hx.Signal(false), hx.Signal(true)
	plan := hx.Signal("free")
	volume := hx.Signal("40")
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
	return coat.Grid(2, 6,
		hx.Form().Class("flex", "flex-col", "gap-4").Body(
			coat.FormField("Name", "f-name", coat.Input("text", name), nameErrors),
			coat.Field(coat.Label("Email", "f-email"), coat.Input("email", email).Id("f-email"), coat.Description("We never share it.")),
			coat.Field(coat.Label("Kind", "f-kind"), coat.Select(kind, "Choose…", coat.Plain("Person", "Company")).Id("f-kind")),
			coat.Field(coat.Label("Bio", "f-bio"), coat.Textarea(bio).Id("f-bio")),
			coat.Checkbox("I agree", agree),
			coat.Switch("Notifications", notify),
			coat.RadioGroup("plan", plan, []coat.Option{{Value: "free", Label: "Free"}, {Value: "pro", Label: "Pro"}}),
			coat.Field(coat.Label("Volume", "f-vol"), coat.Slider(0, 100, 5, volume).Id("f-vol")),
			coat.Fieldset("Search", coat.InputGroup(coat.Input("search", hx.Signal("")).Placeholder("Search…"), coat.With(coat.Addon(coat.Button("Go", coat.Ghost, coat.Sm)), coat.End))),
			coat.FieldSeparator("or"),
			coat.FileInput(nil),
			coat.Row(2, coat.SubmitButton("Submit"), coat.Button("Cancel", coat.Outline))),
		coat.Card(coat.CardHeader("Live values", "Everything above is bound to a signal", nil), coat.CardSection(hx.Pre().Class("text-sm", "whitespace-pre-wrap").BindText(summary))))
}

func navigation() hx.INode {
	tab, line := hx.Signal("account"), hx.Signal("one")
	page := hx.Signal(3)
	pages := hx.Signal(12)
	return coat.Stack(4,
		coat.Breadcrumb(coat.Crumb{Label: "Home", Href: "/"}, coat.Crumb{Label: "Components", Href: "/coat"}, coat.Crumb{Label: "Breadcrumb"}),
		coat.Tabs(tab,
			coat.Tab{ID: "account", Label: "Account", Panel: func() hx.INode { return hx.P().Text("Account settings.") }},
			coat.Tab{ID: "password", Label: "Password", Panel: func() hx.INode { return hx.P().Text("Change your password.") }}),
		coat.LineTabs(line,
			coat.Tab{ID: "one", Label: "One", Panel: func() hx.INode { return hx.P().Text("First.") }},
			coat.Tab{ID: "two", Label: "Two", Panel: func() hx.INode { return hx.P().Text("Second.") }}),
		coat.Accordion("faq",
			coat.AccordionItem{Title: "Is it accessible?", Body: hx.P().Text("Yes: it is the native details element."), Open: true},
			coat.AccordionItem{Title: "Does it need JavaScript?", Body: hx.P().Text("No, only Hix signals.")}),
		coat.Pagination(coat.PaginationOptions{Page: page, Pages: pages}))
}

func dialogs() hx.INode {
	dialog, alert, drawer := hx.Signal(false), hx.Signal(false), hx.Signal(false)
	closeAll := func() { dialog.Set(false); alert.Set(false); drawer.Set(false) }
	return coat.Row(2,
		coat.Button("Open dialog", coat.Outline).OnClick(func(hx.EventContext) { dialog.Set(true) }),
		coat.Button("Open alert dialog", coat.Destructive).OnClick(func(hx.EventContext) { alert.Set(true) }),
		coat.Button("Open drawer", coat.Outline).OnClick(func(hx.EventContext) { drawer.Set(true) }),
		coat.Dialog(coat.DialogOptions{Open: dialog, Title: "Edit profile", Description: "Make changes to your profile here.",
			Body: func() hx.INode { return hx.P().Text("A dialog is in the tree only while it is open.") },
			Footer: func() hx.INode {
				return coat.Button("Save changes").OnClick(func(hx.EventContext) { dialog.Set(false) })
			}}),
		coat.AlertDialog(coat.AlertDialogOptions{Open: alert, Title: "Are you absolutely sure?", Description: "This action cannot be undone.",
			Footer: func() hx.INode {
				return hx.Noop().Body(coat.Button("Cancel", coat.Outline).OnClick(func(hx.EventContext) { closeAll() }),
					coat.Button("Continue").OnClick(func(hx.EventContext) { closeAll() }))
			}}),
		coat.Drawer(coat.DrawerOptions{Open: drawer, Title: "Move goal", Description: "Set your daily activity goal.",
			Body:   func() hx.INode { return hx.P().Text("A drawer comes from the bottom by default.") },
			Footer: func() hx.INode { return coat.Button("Submit").OnClick(func(hx.EventContext) { drawer.Set(false) }) }}))
}

func menus() hx.INode {
	menu, pop := hx.Signal(false), hx.Signal(false)
	show := hx.Signal(true)
	pick, combo := hx.Signal("b"), hx.Signal("")
	query := hx.Signal("")
	chosen := hx.Signal("none")
	fruit := coat.Plain("Apple", "Banana", "Blueberry", "Grapes", "Pineapple")
	return coat.Stack(6,
		coat.Row(3,
			coat.DropdownMenu(menu, coat.Button("Open menu", coat.Outline), []hx.INode{
				coat.MenuHeading("My account"),
				coat.MenuItem("Profile", func() { chosen.Set("profile") }),
				coat.MenuItem("Settings", func() { chosen.Set("settings") }),
				coat.MenuCheckbox("Show toolbar", show),
				coat.MenuSeparator(),
				coat.MenuItem("Delete", func() { chosen.Set("delete") }, coat.Destructive)}),
			coat.Popover(pop, coat.Button("Open popover", coat.Outline), hx.Div().Class("p-2").Body(hx.P().Text("Dimensions"), hx.P().Class("text-sm", "text-muted-foreground").Text("Set the size of the layer."))),
			coat.SelectMenu(pick, "Pick a letter", []coat.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}}),
			coat.Combobox(coat.ComboboxOptions{Value: combo, Placeholder: "Select fruit…", SearchPlaceholder: "Search fruit…", Options: fruit}),
			coat.Tooltip(coat.Button("Hover me", coat.Outline), "Add to library")),
		hx.P().Class("text-sm").Body(hx.Span().Text("Chose: "), hx.Strong().BindText(chosen)),
		coat.Command(query, "Type a command or search…",
			coat.CommandGroup{Heading: "Suggestions", Items: []coat.CommandItem{
				{Label: "Calendar", OnSelect: func() { chosen.Set("calendar") }},
				{Label: "Search emoji"},
				{Label: "Calculator", Shortcut: "⌘K"}}},
			coat.CommandGroup{Heading: "Settings", Items: []coat.CommandItem{{Label: "Profile", Shortcut: "⌘P"}, {Label: "Billing"}}}).Class("max-w-md", "border", "rounded-lg"))
}

func data() hx.INode {
	progress := hx.Signal(40)
	toasts := hx.Signal([]coat.Toast{})
	next := 0
	return coat.Stack(4,
		coat.TableContainer(coat.SimpleTable([]string{"Invoice", "Status", "Amount"}, row("INV001", "Paid", "$250.00"), row("INV002", "Pending", "$150.00"), row("INV003", "Unpaid", "$350.00"))),
		coat.ProgressOf(progress),
		coat.Row(2,
			coat.Button("-10", coat.Outline, coat.Sm).OnClick(func(hx.EventContext) { progress.Set(max(0, progress.Get()-10)) }),
			coat.Button("+10", coat.Outline, coat.Sm).OnClick(func(hx.EventContext) { progress.Set(min(100, progress.Get()+10)) })),
		coat.Row(3, coat.Skeleton("3rem", "3rem", coat.Circle), coat.Stack(2, coat.Skeleton("16rem", "1rem"), coat.Skeleton("12rem", "1rem"))),
		coat.Row(2, coat.Button("Show toast", coat.Outline).OnClick(func(hx.EventContext) {
			next++
			toasts.Set(append(append([]coat.Toast(nil), toasts.Get()...), coat.Toast{ID: next, Title: "Event created", Description: "Sunday, December 03 at 9:00 AM"}))
		})),
		coat.Toaster(toasts, func(id int) {
			var kept []coat.Toast
			for _, t := range toasts.Get() {
				if t.ID != id {
					kept = append(kept, t)
				}
			}
			toasts.Set(kept)
		}))
}

func row(cells ...string) hx.INode {
	nodes := make([]hx.INode, len(cells))
	for i, c := range cells {
		nodes[i] = hx.Td().Text(c)
	}
	return hx.Tr().BodyList(nodes)
}
