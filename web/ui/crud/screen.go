package crud

import (
	"context"

	"github.com/deltegui/hx"

	"owldemo/web/ui/bs"
)

// Column is one column of the table. Cell gets the item as something that can
// change: the row is reused when the list reloads, so the cell has to follow it.
type Column[T any] struct {
	Header string // the key of its text
	Cell   func(item hx.Gettable[T], t Texts) hx.INode
}

// TextCol shows a string.
func TextCol[T any](header string, get func(T) string) Column[T] {
	return Column[T]{header, func(item hx.Gettable[T], _ Texts) hx.INode {
		return hx.Span().BindText(hx.Computed(func() string { return get(item.Get()) }))
	}}
}

// ActiveCol shows a green or grey badge.
func ActiveCol[T any](header string, get func(T) bool) Column[T] {
	return Column[T]{header, func(item hx.Gettable[T], t Texts) hx.INode {
		badge := hx.Span().Class("badge")
		badge.Effect(func() {
			on := get(item.Get())
			badge.ClassIf("text-bg-success", on).ClassIf("text-bg-secondary", !on)
			if on {
				badge.Text(t.Get("Shared.Active"))
			} else {
				badge.Text(t.Get("Shared.Disabled"))
			}
		})
		return badge
	}}
}

// Config describes a screen. The API is the one the handlers of its feature give.
type Config[T any] struct {
	Texts   Texts // what the screen says (see Texts)
	API     string
	ID      func(T) int
	Name    func(T) string // how the confirmation dialog calls an item
	Active  func(T) bool
	Columns []Column[T]
	Filters []Field
	Form    []Field
	// Initial is the first page, when the server rendered it. The screen starts
	// from it and does not load it again. It can be nil: the screen loads it.
	Initial *Page[T]
	// Changed is called after an item is saved, deleted, enabled or disabled, so
	// other islands that show the same data can read it again. It can be nil.
	Changed func()
}

// listQuery is what the list depends on. Changing it loads another page.
type listQuery struct {
	Filters string // already encoded: name=a&status=active&
	Page    int
}

type saveArgs struct {
	ID   int
	Body map[string]any
}

// Screen builds the whole thing: filter, paginated list, the form in a modal and
// the confirmations. All the state is in signals, and the server only answers JSON.
func Screen[T any](cfg Config[T]) hx.INode {
	api := Client[T]{Base: cfg.API}
	t := cfg.Texts
	notice, failure := hx.Signal(""), hx.Signal("")

	// The list and its filter. The filter is typed in `drafts`; "Search" copies it to `applied`.
	drafts := map[string]*hx.SignalT[string]{}
	for _, f := range cfg.Filters {
		drafts[f.Key] = hx.Signal("")
	}
	applied := hx.Signal(listQuery{Page: 1})
	loadPage := func(ctx context.Context, q listQuery) (Page[T], error) {
		return api.List(ctx, q.Filters, q.Page)
	}
	var list *hx.ResourceT[Page[T]]
	if cfg.Initial != nil {
		list = hx.ResourceForWith(*cfg.Initial, applied, loadPage)
	} else {
		list = hx.ResourceFor(applied, loadPage)
	}
	current := func() Page[T] { return list.Value().Get() }
	rows := getter[[]T](func() []T { return current().Items })
	page := hx.Computed(func() int { return max(1, current().Page) })
	pages := hx.Computed(func() int { return max(1, current().Pages) })
	total := hx.Computed(func() int { return current().Total })

	search := func() {
		query := ""
		for _, f := range cfg.Filters {
			if v := drafts[f.Key].Get(); v != "" {
				query += f.Key + "=" + Escape(v) + "&"
			}
		}
		applied.Set(listQuery{Filters: query, Page: 1})
	}
	clear := func() {
		for _, draft := range drafts {
			draft.Set("")
		}
		search()
	}
	goTo := func(p int) {
		q := applied.Get()
		q.Page = p
		applied.Set(q)
	}

	// Create and edit.
	form := newFormState(cfg.Form)
	editing := hx.Signal(0) // the id being edited; 0 is a new one
	formOpen := hx.Signal(false)
	save := hx.Action(func(ctx context.Context, a saveArgs) (T, error) { return api.Save(ctx, a.ID, a.Body) })
	changed := func() {
		list.Refetch()
		if cfg.Changed != nil {
			cfg.Changed()
		}
	}
	save.OnSuccess(func(_ T, a saveArgs) {
		formOpen.Set(false)
		notice.Set(t.Get("SavedMessage"))
		changed()
	})
	save.OnError(func(err error, _ saveArgs) {
		if byField, ok := FieldErrors(err); ok {
			form.errors.Set(byField)
			return
		}
		form.other.Set(err.Error())
	})
	submit := func() {
		form.errors.Set(map[string][]string{})
		form.other.Set("")
		save.Run(saveArgs{ID: editing.Get(), Body: form.body()})
	}
	openForm := func(item *T) {
		notice.Set("")
		if item == nil {
			editing.Set(0)
			form.load(nil)
		} else {
			editing.Set(cfg.ID(*item))
			form.load(*item)
		}
		formOpen.Set(true)
	}

	// Delete, with a confirmation.
	deleting := hx.Signal(0)
	deletingName := hx.Signal("")
	remove := hx.Action(func(ctx context.Context, id int) (struct{}, error) { return struct{}{}, api.Delete(ctx, id) })
	remove.OnSuccess(func(_ struct{}, _ int) {
		deleting.Set(0)
		notice.Set(t.Get("DeletedMessage"))
		changed()
	})
	remove.OnError(func(err error, _ int) {
		deleting.Set(0)
		failure.Set(err.Error())
	})
	deleteOpen := hx.Computed(func() bool { return deleting.Get() != 0 })

	// Enable and disable.
	toggle := hx.Action(func(ctx context.Context, id int) (T, error) { return api.Toggle(ctx, id) })
	toggle.OnSuccess(func(_ T, _ int) { changed() })
	toggle.OnError(func(err error, _ int) { failure.Set(err.Error()) })

	headers := make([]string, 0, len(cfg.Columns)+1)
	for _, c := range cfg.Columns {
		headers = append(headers, t.Get(c.Header))
	}
	headers = append(headers, "")

	table := hx.TBody().Body(hx.EachKeyed(rows, cfg.ID,
		func(id int, item hx.Gettable[T], _ hx.Gettable[int]) hx.INode {
			cells := make([]hx.INode, 0, len(cfg.Columns)+1)
			for _, c := range cfg.Columns {
				cells = append(cells, hx.Td().Body(c.Cell(item, t)))
			}
			toggleLabel := hx.Computed(func() string {
				if cfg.Active(item.Get()) {
					return t.Get("Shared.Disable")
				}
				return t.Get("Shared.Enable")
			})
			cells = append(cells, hx.Td().Class("text-end", "text-nowrap").Body(
				bs.SmallButton("outline-primary", t.Get("Shared.Edit")).OnClick(func(hx.EventContext) {
					it := hx.UntrackGet(item)
					openForm(&it)
				}),
				hx.Span().Text(" "),
				bs.SmallButton("outline-secondary", "").BindText(toggleLabel).OnClick(func(hx.EventContext) {
					failure.Set("")
					toggle.Run(cfg.ID(hx.UntrackGet(item)))
				}),
				hx.Span().Text(" "),
				bs.SmallButton("outline-danger", t.Get("Shared.Delete")).OnClick(func(hx.EventContext) {
					it := hx.UntrackGet(item)
					failure.Set("")
					deletingName.Set(cfg.Name(it))
					deleting.Set(cfg.ID(it))
				})))
			return hx.Tr().BodyList(cells)
		}))

	status := hx.Span().Class("text-muted", "small")
	status.Effect(func() {
		switch {
		case list.Loading().Get():
			status.Text(t.Get("Shared.Loading"))
		case list.Error().Get() != nil:
			status.Text(t.Get("Shared.LoadError", list.Error().Get().Error()))
		case total.Get() == 0:
			status.Text(t.Get("Shared.NothingFound"))
		default:
			status.Text(t.Get("Shared.Found", total.Get()))
		}
	})

	return hx.Div().Body(
		bs.Row(
			bs.Col(8, hx.H1().Text(t.Get("Title")), hx.P().Class("text-muted").Text(t.Get("Subtitle"))),
			bs.Col(4, bs.Button("primary", t.Get("NewButton")).Class("float-end").OnClick(func(hx.EventContext) { openForm(nil) }))),
		bs.Alert("success", notice, func() { notice.Set("") }),
		bs.Alert("danger", failure, func() { failure.Set("") }),
		bs.Card(t.Get("Shared.Filter"), filterForm(t, cfg.Filters, drafts, search, clear)),
		bs.Card("", bs.Table(headers, table), hx.Div().Class("d-flex", "justify-content-between", "align-items-center").Body(
			status, bs.Pagination(page, pages, goTo))),
		bs.Modal(formOpen, t.Get("FormTitle"), func() hx.INode { return formBody(cfg, form, editing, submit) },
			func() hx.INode {
				return hx.Div().Body(
					bs.Button("secondary", t.Get("Shared.Cancel")).OnClick(func(hx.EventContext) { formOpen.Set(false) }),
					hx.Span().Text(" "),
					saveButton(t, save.Pending(), submit))
			}, func() { formOpen.Set(false) }),
		bs.Modal(deleteOpen, t.Get("DeleteTitle"), func() hx.INode {
			return hx.P().BindText(getter[string](func() string { return t.Get("DeleteQuestion", deletingName.Get()) }))
		}, func() hx.INode {
			return hx.Div().Body(
				bs.Button("secondary", t.Get("Shared.Cancel")).OnClick(func(hx.EventContext) { deleting.Set(0) }),
				hx.Span().Text(" "),
				bs.Button("danger", t.Get("Shared.Delete")).OnClick(func(hx.EventContext) { remove.Run(hx.UntrackGet(deleting)) }))
		}, func() { deleting.Set(0) }))
}

func saveButton(t Texts, pending hx.Gettable[bool], submit func()) hx.INode {
	button := bs.Button("primary", t.Get("Shared.Save"))
	button.Effect(func() {
		if pending.Get() {
			button.Text(t.Get("Shared.Saving")).Attribute("disabled", "disabled")
		} else {
			button.Text(t.Get("Shared.Save")).RemoveAttribute("disabled")
		}
	})
	return button.OnClick(func(hx.EventContext) { submit() })
}

func filterForm(t Texts, fields []Field, drafts map[string]*hx.SignalT[string], search, clear func()) hx.INode {
	cols := make([]hx.INode, 0, len(fields))
	for _, f := range fields {
		var control hx.INode
		if f.Kind == Select {
			control = bs.Select(drafts[f.Key], t.Get("Shared.All"), options(t, f))
		} else {
			control = bs.TextInput(drafts[f.Key])
		}
		cols = append(cols, bs.Col(4, bs.Field(t.Get(f.Label), control, nil)))
	}
	return hx.Form().OnSubmit(func(ctx hx.EventContext) {
		bs.PreventDefault(ctx)
		search()
	}).Body(
		bs.Row(cols...),
		bs.SubmitButton("primary", t.Get("Shared.Search")),
		hx.Span().Text(" "),
		bs.Button("outline-secondary", t.Get("Shared.Clear")).OnClick(func(hx.EventContext) { clear() }))
}

func formBody[T any](cfg Config[T], form *formState, editing hx.Gettable[int], submit func()) hx.INode {
	controls := make([]hx.INode, 0, len(cfg.Form)+1)
	t := cfg.Texts
	controls = append(controls, bs.Alert("danger", form.other, nil))
	for _, f := range cfg.Form {
		switch f.Kind {
		case Check:
			controls = append(controls, bs.Check(form.checks[f.Key], t.Get(f.Label)))
		case Select:
			controls = append(controls, bs.Field(t.Get(f.Label), bs.Select(form.text[f.Key], t.Get("Shared.Choose"), options(t, f)), form.errorsOf(f.Key)))
		case Number:
			controls = append(controls, bs.Field(t.Get(f.Label), bs.NumberInput(form.text[f.Key]), form.errorsOf(f.Key)))
		default:
			controls = append(controls, bs.Field(t.Get(f.Label), bs.TextInput(form.text[f.Key]), form.errorsOf(f.Key)))
		}
	}
	return hx.Form().OnSubmit(func(ctx hx.EventContext) {
		bs.PreventDefault(ctx)
		submit()
	}).BodyList(controls)
}

// options are the choices of a Select field, with the text of each.
func options(t Texts, f Field) []bs.Option {
	out := make([]bs.Option, len(f.Options))
	for i, value := range f.Options {
		out[i] = bs.Option{Value: value, Label: t.Or("Opt."+value, value)}
	}
	return out
}
