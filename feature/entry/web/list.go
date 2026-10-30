package entryweb

import (
	"context"
	"strconv"

	"github.com/deltegui/hx"

	"owldemo/web/ui/bs"
)

// column is a header of the list. Key is what the server sorts by, empty if it does not.
type column struct{ Key, Label string }

var columns = []column{
	{"id", "Col.ID"},
	{"carriedOutAt", "Col.Date"},
	{"operation", "Col.Operation"},
	{"samount", "Col.SAmount"},
	{"", "Col.Amount"},
	{"category", "Col.Category"},
	{"", "Col.Tags"},
}

const defaultOrderBy = "carriedOutAt"

// list is the filter and the paginated, sortable table of entries. initial is
// the first page, when the server rendered it.
func list(a *app, initial *Listing) hx.INode {
	t := a.t

	// The filter is typed in the drafts; Search copies it to the query.
	concept, category, operation := hx.Signal(""), hx.Signal(""), hx.Signal("")
	from, to := hx.Signal(""), hx.Signal("")
	applied := hx.Signal(Query{Page: 1, OrderBy: defaultOrderBy, Order: "desc"})

	var results *hx.ResourceT[Listing]
	if initial != nil {
		results = hx.ResourceForWith(*initial, applied, fetchList)
	} else {
		results = hx.ResourceFor(applied, fetchList)
	}
	a.reload = results.Refetch
	current := func() Listing { return results.Value().Get() }
	rows := getter[[]Row](func() []Row { return current().Items })
	page := hx.Computed(func() int { return max(1, current().Page) })
	pages := hx.Computed(func() int { return max(1, current().Pages) })
	total := hx.Computed(func() int { return current().Total })
	sum := hx.Computed(func() string { return money(current().Sum) })

	search := func() {
		q := hx.UntrackGet(applied)
		applied.Set(Query{
			Concept: hx.UntrackGet(concept), Category: hx.UntrackGet(category), Operation: hx.UntrackGet(operation),
			From: hx.UntrackGet(from), To: hx.UntrackGet(to),
			OrderBy: q.OrderBy, Order: q.Order, Page: 1,
		})
	}
	clear := func() {
		for _, draft := range []*hx.SignalT[string]{concept, category, operation, from, to} {
			draft.Set("")
		}
		search()
	}
	goTo := func(p int) {
		q := hx.UntrackGet(applied)
		q.Page = p
		applied.Set(q)
	}
	// Clicking a header sorts by it; clicking it again turns the order around.
	sortBy := func(key string) {
		q := hx.UntrackGet(applied)
		if q.OrderBy == key && q.Order == "asc" {
			q.Order = "desc"
		} else {
			q.OrderBy, q.Order = key, "asc"
		}
		q.Page = 1
		applied.Set(q)
	}

	// Delete, with a confirmation.
	deleting := hx.Signal(0)
	remove := hx.Action(func(ctx context.Context, id int) (struct{}, error) { return struct{}{}, deleteEntry(ctx, id) })
	remove.OnSuccess(func(_ struct{}, id int) {
		deleting.Set(0)
		a.notice.Set(t.Get("DeletedMessage", id))
		results.Refetch()
	})
	remove.OnError(func(err error, _ int) {
		deleting.Set(0)
		a.failure.Set(err.Error())
	})
	deleteOpen := hx.Computed(func() bool { return deleting.Get() != 0 })

	head := make([]hx.INode, 0, len(columns)+1)
	for _, c := range columns {
		th := hx.Th().Attribute("scope", "col")
		if c.Key == "" {
			head = append(head, th.Text(t.Get(c.Label)))
			continue
		}
		c := c
		th.Class("user-select-none").Style("cursor", "pointer").Attribute("data-sort", c.Key)
		th.Effect(func() {
			q := applied.Get()
			label := t.Get(c.Label)
			switch {
			case q.OrderBy != c.Key:
			case q.Order == "asc":
				label += " ▲"
			default:
				label += " ▼"
			}
			th.Text(label)
		})
		head = append(head, th.OnClick(func(hx.EventContext) { sortBy(c.Key) }))
	}
	head = append(head, hx.Th())

	body := hx.TBody().Body(hx.EachKeyed(rows, func(r Row) int { return r.ID },
		func(id int, item hx.Gettable[Row], _ hx.Gettable[int]) hx.INode {
			text := func(get func(Row) string) hx.INode {
				return hx.Span().BindText(hx.Computed(func() string { return get(item.Get()) }))
			}
			return hx.Tr().Body(
				hx.Td().Body(text(func(r Row) string { return strconv.Itoa(r.ID) })),
				hx.Td().Class("text-nowrap").Body(text(func(r Row) string { return r.CarriedOutAt })),
				hx.Td().Body(text(func(r Row) string {
					if r.Operation == "" {
						return t.Get("BasicOperationNotAvailable")
					}
					return t.Get(r.Operation)
				})),
				hx.Td().Class("text-end", "text-nowrap").Body(text(func(r Row) string { return money(r.SAmount) + " €" })),
				hx.Td().Class("text-end", "text-nowrap").Body(text(func(r Row) string { return money(r.Amount) + " " + r.Sign })),
				hx.Td().Body(text(func(r Row) string { return r.Category })),
				hx.Td().Body(text(func(r Row) string { return r.Tags })),
				hx.Td().Class("text-end", "text-nowrap").Body(
					bs.SmallButton("outline-secondary", t.Get("Shared.View")).OnClick(func(hx.EventContext) {
						a.show("view", hx.UntrackGet(item).ID)
					}),
					hx.Span().Text(" "),
					bs.SmallButton("outline-primary", t.Get("Shared.Edit")).OnClick(func(hx.EventContext) {
						a.show("edit", hx.UntrackGet(item).ID)
					}),
					hx.Span().Text(" "),
					bs.SmallButton("outline-danger", t.Get("Shared.Delete")).OnClick(func(hx.EventContext) {
						a.failure.Set("")
						deleting.Set(hx.UntrackGet(item).ID)
					})))
		}))

	status := hx.Span().Class("text-muted", "small")
	status.Effect(func() {
		switch {
		case results.Loading().Get():
			status.Text(t.Get("Shared.Loading"))
		case results.Error().Get() != nil:
			status.Text(t.Get("Shared.LoadError", results.Error().Get().Error()))
		case total.Get() == 0:
			status.Text(t.Get("Shared.NothingFound"))
		default:
			status.Text(t.Get("Shared.Found", total.Get()))
		}
	})

	categories := make([]bs.Option, 0, len(a.look.Categories))
	for _, c := range a.look.Categories {
		categories = append(categories, bs.Option{Value: strconv.Itoa(c.ID), Label: c.Name})
	}
	operations := make([]bs.Option, 0, len(a.look.Operations))
	for _, op := range a.look.Operations {
		operations = append(operations, bs.Option{Value: strconv.Itoa(op.ID), Label: t.Get(op.Name)})
	}
	filter := hx.Form().OnSubmit(func(ctx hx.EventContext) {
		bs.PreventDefault(ctx)
		search()
	}).Body(
		bs.Row(
			bs.Col(3, bs.Field(t.Get("Filter.Concept"), bs.TextInput(concept), nil)),
			bs.Col(3, bs.Field(t.Get("Filter.Category"), bs.Select(category, t.Get("Shared.All"), categories), nil)),
			bs.Col(3, bs.Field(t.Get("Filter.Operation"), bs.Select(operation, t.Get("Shared.All"), operations), nil)),
			bs.Col(3, bs.Field(t.Get("Filter.From"), bs.DateInput(from), nil)),
			bs.Col(3, bs.Field(t.Get("Filter.To"), bs.DateInput(to), nil))),
		bs.SubmitButton("primary", t.Get("Shared.Search")),
		hx.Span().Text(" "),
		bs.Button("outline-secondary", t.Get("Shared.Clear")).OnClick(func(hx.EventContext) { clear() }))

	table := hx.Div().Class("table-responsive").Body(
		hx.Table().Class("table", "table-striped", "table-hover", "align-middle").Body(
			hx.THead().Body(hx.Tr().BodyList(head)),
			body,
			hx.TFoot().Body(hx.Tr().Class("fw-bold").Body(
				hx.Td().Attribute("colspan", "3").Class("text-end").Text(t.Get("Total")+":"),
				hx.Td().Class("text-end").Body(hx.Span().BindText(hx.Computed(func() string { return sum.Get() + " €" }))),
				hx.Td().Attribute("colspan", "4")))))

	return hx.Div().Body(
		bs.Row(
			bs.Col(8, hx.H1().Text(t.Get("Title")), hx.P().Class("text-muted").Text(t.Get("Subtitle"))),
			bs.Col(4, bs.Button("primary", t.Get("NewButton")).Class("float-end").OnClick(func(hx.EventContext) { a.show("edit", 0) }))),
		bs.Alert("success", a.notice, func() { a.notice.Set("") }),
		bs.Alert("danger", a.failure, func() { a.failure.Set("") }),
		bs.Card(t.Get("Shared.Filter"), filter),
		bs.Card("", table, hx.Div().Class("d-flex", "justify-content-between", "align-items-center").Body(
			status, bs.Pagination(page, pages, goTo))),
		bs.Modal(deleteOpen, t.Get("DeleteTitle"), func() hx.INode {
			return hx.P().BindText(getter[string](func() string { return t.Get("DeleteQuestion", deleting.Get()) }))
		}, func() hx.INode {
			return hx.Div().Body(
				bs.Button("secondary", t.Get("Shared.Cancel")).OnClick(func(hx.EventContext) { deleting.Set(0) }),
				hx.Span().Text(" "),
				bs.Button("danger", t.Get("Shared.Delete")).OnClick(func(hx.EventContext) { remove.Run(hx.UntrackGet(deleting)) }))
		}, func() { deleting.Set(0) }))
}
