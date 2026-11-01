// Package bs wraps Bootstrap 5 in Hix nodes. It uses only classes, no Bootstrap
// JavaScript (a modal is a node that is in the tree or not), and it builds on
// the server and in the browser.
package bs

import (
	"strconv"
	"strings"

	"github.com/deltegui/hx"
)

func Row(children ...hx.INode) hx.INode { return hx.Div().Class("row").Body(children...) }

func Col(span int, children ...hx.INode) hx.INode {
	return hx.Div().Class("col-12", "col-md-"+strconv.Itoa(span)).Body(children...)
}

func Card(title string, body ...hx.INode) hx.INode {
	card := hx.Div().Class("card", "mb-3")
	if title != "" {
		card.Body(hx.Div().Class("card-header").Text(title), hx.Div().Class("card-body").Body(body...))
	} else {
		card.Body(hx.Div().Class("card-body").Body(body...))
	}
	return card
}

// Button is a button that does not submit. variant is primary, secondary, danger...
func Button(variant, label string) hx.INode {
	return hx.Button().Type("button").Class("btn", "btn-"+variant).Text(label)
}

func SmallButton(variant, label string) hx.INode { return Button(variant, label).Class("btn-sm") }

// SubmitButton submits the form it is in.
func SubmitButton(variant, label string) hx.INode {
	return hx.Button().Type("submit").Class("btn", "btn-"+variant).Text(label)
}

func Spinner() hx.INode {
	return hx.Div().Class("spinner-border", "spinner-border-sm").Attribute("role", "status")
}

// Alert shows text while it is not empty. dismiss, if not nil, adds a close button.
func Alert(variant string, text hx.Gettable[string], dismiss func()) hx.INode {
	return hx.Div().Body(hx.Dynamic(func() hx.INode {
		message := text.Get()
		if message == "" {
			return nil
		}
		content := []hx.INode{hx.Span().Text(message)}
		if dismiss != nil {
			content = append(content, hx.Button().Type("button").Class("btn-close").Attribute("aria-label", "Close").OnClick(func(hx.EventContext) { dismiss() }))
		}
		return hx.Div().Class("alert", "alert-"+variant, "d-flex", "justify-content-between", "align-items-center").Attribute("role", "alert").BodyList(content)
	}))
}

// Table is a striped table with a header; put the rows in body (a TBody).
func Table(headers []string, body hx.INode) hx.INode {
	cells := make([]hx.INode, 0, len(headers))
	for _, h := range headers {
		cells = append(cells, hx.Th().Attribute("scope", "col").Text(h))
	}
	row := hx.Tr().BodyList(cells)
	return hx.Div().Class("table-responsive").Body(
		hx.Table().Class("table", "table-striped", "table-hover", "align-middle").Body(hx.THead().Body(row), body))
}

// Pagination shows previous, a window of pages around the current one, and next.
func Pagination(page, pages hx.Gettable[int], goTo func(int)) hx.INode {
	item := func(label string, target int, enabled, current bool) hx.INode {
		button := hx.Button().Type("button").Class("page-link").Text(label)
		if enabled && !current {
			button.OnClick(func(hx.EventContext) { goTo(target) })
		}
		return hx.Li().Class("page-item").ClassIf("active", current).ClassIf("disabled", !enabled).Body(button)
	}
	list := hx.Ul().Class("pagination", "pagination-sm", "mb-0").Body(hx.Dynamic(func() hx.INode {
		current, last := page.Get(), pages.Get()
		items := []hx.INode{item("«", current-1, current > 1, false)}
		for p := max(1, current-2); p <= min(last, current+2); p++ {
			items = append(items, item(strconv.Itoa(p), p, true, p == current))
		}
		items = append(items, item("»", current+1, current < last, false))
		return hx.Noop().BodyList(items)
	}))
	return hx.Nav().Attribute("aria-label", "Pages").Body(list)
}

// Modal is in the tree while open is true. The body and the footer are built
// each time it opens, so a closed modal costs nothing.
func Modal(open hx.Gettable[bool], title string, body, footer func() hx.INode, onClose func()) hx.INode {
	return hx.Div().Body(hx.Dynamic(func() hx.INode {
		if !open.Get() {
			return nil
		}
		var dialog hx.INode
		hx.Untrack(func() {
			dialog = hx.Div().Class("modal", "show", "d-block").Attribute("tabindex", "-1").Attribute("role", "dialog").Body(
				hx.Div().Class("modal-dialog").Body(
					hx.Div().Class("modal-content").Body(
						hx.Div().Class("modal-header").Body(
							hx.H5().Class("modal-title").Text(title),
							hx.Button().Type("button").Class("btn-close").Attribute("aria-label", "Close").OnClick(func(hx.EventContext) { onClose() })),
						hx.Div().Class("modal-body").Body(body()),
						hx.Div().Class("modal-footer").Body(footer()))))
		})
		return hx.Noop().Body(dialog, hx.Div().Class("modal-backdrop", "show"))
	}))
}

// Field is a label, a control and the errors of that field.
func Field(label string, control hx.INode, errs hx.Gettable[[]string]) hx.INode {
	messages := hx.Div().Class("invalid-feedback", "d-block")
	if errs != nil {
		messages.Effect(func() { messages.Text(strings.Join(errs.Get(), " · ")) })
	}
	return hx.Div().Class("mb-3").Body(hx.Label().Class("form-label").Text(label), control, messages)
}

func TextInput(value *hx.SignalT[string]) hx.INode {
	return hx.Input().Type("text").Class("form-control").BindValue(value).BindOnInput(value)
}

// DateInput is an input for a day, as "2026-03-01".
func DateInput(value *hx.SignalT[string]) hx.INode {
	return hx.Input().Type("date").Class("form-control").BindValue(value).BindOnInput(value)
}

// TextArea is a multi line text.
func TextArea(value *hx.SignalT[string]) hx.INode {
	return hx.TextArea().Class("form-control").Attribute("rows", "2").BindValue(value).BindOnInput(value)
}

func NumberInput(value *hx.SignalT[string]) hx.INode {
	return hx.Input().Type("number").Attribute("step", "any").Class("form-control").BindValue(value).BindOnInput(value)
}

// Option is an item of a Select: what it sends and what the user reads.
type Option struct{ Value, Label string }

// Plain makes options whose label is their value.
func Plain(values ...string) []Option {
	options := make([]Option, len(values))
	for i, v := range values {
		options[i] = Option{v, v}
	}
	return options
}

// Select has an empty first option with the text of blank.
func Select(value *hx.SignalT[string], blank string, options []Option) hx.INode {
	sel := hx.Select().Class("form-select")
	nodes := []hx.INode{hx.Option().Value("").Text(blank)}
	for _, o := range options {
		nodes = append(nodes, hx.Option().Value(o.Value).Text(o.Label))
	}
	sel.BodyList(nodes)
	return sel.BindValue(value).BindOnChange(value)
}

func Check(value *hx.SignalT[bool], label string) hx.INode {
	return hx.Div().Class("form-check", "mb-3").Body(
		hx.Label().Class("form-check-label").Body(
			hx.Input().Type("checkbox").Class("form-check-input").BindChecked(value).BindOnChecked(value),
			hx.Span().Text(" "+label)))
}
