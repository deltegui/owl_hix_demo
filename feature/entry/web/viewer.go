package entryweb

import (
	"strconv"
	"strings"

	"github.com/deltegui/hx"

	"owldemo/sys/currency"
	"owldemo/sys/ledger"
	"owldemo/web/ui/bs"
)

// viewer is the entry read only, for who cannot or does not want to change it.
func viewer(a *app, e Entry) hx.INode {
	t := a.t

	tagNames := []string{}
	for _, n := range a.look.Tags {
		for _, id := range e.TagIds {
			if id == n.ID {
				tagNames = append(tagNames, n.Name)
			}
		}
	}
	item := func(label, value string) hx.INode {
		return hx.Div().Class("row", "mb-2").Body(
			hx.Div().Class("col-12", "col-md-3", "text-muted").Text(label),
			hx.Div().Class("col-12", "col-md-9").Attribute("data-view", label).Text(value))
	}

	var what hx.INode
	if e.Mode == ModeBasic {
		op, _ := a.operation(e.OperationId)
		amount := 0.0
		if e.Amount != nil {
			amount = *e.Amount
		}
		what = hx.P().Class("fs-5").Attribute("data-role", "phrase").Text(strings.Join([]string{
			t.Get(op.Name), t.Get("Phrase.Of"), money(amount), currency.Currency(e.Currency).Sign(),
			t.Get("Phrase.From"), a.fullAccountName(e.CreditAccount), t.Get("Phrase.To"), a.fullAccountName(e.DebitAccount),
		}, " "))
	} else {
		what = viewLines(a, e)
	}

	meta := hx.P().Class("text-muted", "small", "mb-0")
	meta.Text(t.Get("CreatedBy", e.CreatedBy, e.CreatedAt) + " · " + t.Get("UpdatedBy", e.UpdatedBy, e.UpdatedAt))

	return hx.Div().Attribute("data-layer", "viewer").Body(
		bs.Row(
			bs.Col(8, hx.H1().Text(t.Get("ViewTitle", e.ID)), hx.P().Class("text-muted").Text(t.Get("ViewSubtitle"))),
			bs.Col(4, hx.Div().Class("float-end").Body(
				bs.Button("primary", t.Get("Shared.Edit")).OnClick(func(hx.EventContext) { a.show("edit", e.ID) }),
				hx.Span().Text(" "),
				bs.Button("secondary", t.Get("Shared.Return")).OnClick(func(hx.EventContext) { a.close() })))),
		bs.Card(t.Get("Operation"), what),
		bs.Card("",
			item(t.Get("CarriedOutAt"), e.Date+" "+twoDigits(e.Hour)+":"+twoDigits(e.Minute)),
			item(t.Get("Category"), a.categoryName(e.CategoryId)),
			item(t.Get("Tags"), strings.Join(tagNames, ", ")),
			item(t.Get("Concept"), e.Concept),
			meta))
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// viewLines is the table of lines with the full name of each account and the
// total of each side.
func viewLines(a *app, e Entry) hx.INode {
	t := a.t
	var debit, credit int64
	rows := make([]hx.INode, 0, len(e.Lines))
	for _, l := range e.Lines {
		var d, c, arrow string
		acc, _ := a.account(l.Account)
		side := ledger.SideDebit
		if l.Debit != nil {
			d = money(*l.Debit)
			debit += cents(*l.Debit, l.Rate)
		}
		if l.Credit != nil {
			c = money(*l.Credit)
			credit += cents(*l.Credit, l.Rate)
			side = ledger.SideCredit
		}
		if ledger.AccountGroup(acc.Group).Variation(side) == ledger.Increase {
			arrow = "↑"
		} else {
			arrow = "↓"
		}
		rows = append(rows, hx.Tr().Attribute("data-line", strconv.Itoa(l.Account)).Body(
			hx.Td().Text(a.fullAccountName(l.Account)+" "+arrow),
			hx.Td().Class("text-end").Text(d), hx.Td().Class("text-end").Text(c),
			hx.Td().Text(l.Currency), hx.Td().Class("text-end").Text(money(l.SAmount)+" €"), hx.Td().Text(l.Notes)))
	}
	return hx.Div().Class("table-responsive").Body(hx.Table().Class("table", "align-middle").Body(
		hx.THead().Body(hx.Tr().Body(
			hx.Th().Text(t.Get("Account")), hx.Th().Class("text-end").Text(t.Get("Debit")), hx.Th().Class("text-end").Text(t.Get("Credit")),
			hx.Th().Text(t.Get("Currency")), hx.Th().Class("text-end").Text(t.Get("InEuros")), hx.Th().Text(t.Get("Notes")))),
		hx.TBody().BodyList(rows),
		hx.TFoot().Body(hx.Tr().Class("fw-bold").Body(
			hx.Td().Class("text-end").Text(t.Get("Total")+":"),
			hx.Td().Class("text-end").Text(money(float64(debit)/100)+" €"),
			hx.Td().Class("text-end").Text(money(float64(credit)/100)+" €"),
			hx.Td().Attribute("colspan", "3")))))
}
