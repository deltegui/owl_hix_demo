package formsweb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/deltegui/hx"

	"owldemo/web/ui/bs"
	"owldemo/web/ui/crud"
)

// Island builds the form from the state the server sent (JSON).
func Island(state string) hx.INode {
	var s State
	if state != "" {
		json.Unmarshal([]byte(state), &s)
	}
	return Build(s)
}

// Build draws the form. There is nothing to load: the description is in the state.
func Build(state State) hx.INode {
	form, t := state.Form, TextsFor(state.Lang)
	text := map[string]*hx.SignalT[string]{}
	checks := map[string]*hx.SignalT[bool]{}
	for _, f := range form.Fields {
		if f.Kind == Check {
			checks[f.Key] = hx.Signal(true)
		} else {
			text[f.Key] = hx.Signal("")
		}
	}
	errs := hx.Signal(map[string][]string{})
	other := hx.Signal("")
	notice := hx.Signal("")

	send := hx.Action(func(ctx context.Context, body map[string]any) (map[string]any, error) {
		return crud.Send[map[string]any](ctx, "POST", form.Endpoint, body)
	})
	send.OnSuccess(func(saved map[string]any, _ map[string]any) {
		notice.Set(t.Get("Saved", saved["id"]))
		for _, f := range form.Fields {
			if f.Kind != Check {
				text[f.Key].Set("")
			}
		}
	})
	send.OnError(func(err error, _ map[string]any) {
		if byField, ok := crud.FieldErrors(err); ok {
			errs.Set(byField)
			return
		}
		other.Set(err.Error())
	})

	options := func(f Field) []bs.Option {
		out := make([]bs.Option, len(f.Options))
		for i, value := range f.Options {
			out[i] = bs.Option{Value: value, Label: t.Or("Opt."+value, value)}
		}
		return out
	}
	controls := []hx.INode{bs.Alert("success", notice, func() { notice.Set("") }), bs.Alert("danger", other, nil)}
	for _, f := range form.Fields {
		errsOf := getter[[]string](func() []string { return errs.Get()[f.Key] })
		label := t.Get(f.Label)
		switch f.Kind {
		case Check:
			controls = append(controls, bs.Check(checks[f.Key], label))
		case Select:
			controls = append(controls, bs.Field(label, bs.Select(text[f.Key], t.Get("Shared.Choose"), options(f)), errsOf))
		case Number:
			controls = append(controls, bs.Field(label, bs.NumberInput(text[f.Key]), errsOf))
		default:
			controls = append(controls, bs.Field(label, bs.TextInput(text[f.Key]), errsOf))
		}
	}
	controls = append(controls, saveButton(t, send.Pending()))

	return bs.Card(t.Get(form.Title), hx.Form().OnSubmit(func(ctx hx.EventContext) {
		bs.PreventDefault(ctx)
		notice.Set("")
		other.Set("")
		body, problems := read(t, form, text, checks)
		errs.Set(problems)
		if len(problems) == 0 {
			send.Run(body)
		}
	}).BodyList(controls))
}

func saveButton(t crud.Texts, pending hx.Gettable[bool]) hx.INode {
	button := bs.SubmitButton("primary", t.Get("Shared.Save"))
	button.Effect(func() {
		if pending.Get() {
			button.Text(t.Get("Shared.Saving")).Attribute("disabled", "disabled")
		} else {
			button.Text(t.Get("Shared.Save")).RemoveAttribute("disabled")
		}
	})
	return button
}

// read builds the body to send and checks the rules of the form. The server
// checks them again: this only saves a trip.
func read(t crud.Texts, form Form, text map[string]*hx.SignalT[string], checks map[string]*hx.SignalT[bool]) (map[string]any, map[string][]string) {
	body := map[string]any{}
	problems := map[string][]string{}
	for _, f := range form.Fields {
		if f.Kind == Check {
			body[f.Key] = checks[f.Key].Get()
			continue
		}
		value := strings.TrimSpace(text[f.Key].Get())
		var size float64 // the number, or the length of the text
		if f.Kind == Number {
			size, _ = strconv.ParseFloat(value, 64)
			body[f.Key] = size
		} else {
			size = float64(len([]rune(value)))
			body[f.Key] = value
		}
		switch {
		case f.Required && value == "":
			problems[f.Key] = []string{t.Get("Shared.Required")}
		case f.Min != nil && size < *f.Min:
			problems[f.Key] = []string{t.Get("Shared.AtLeast", fmt.Sprint(*f.Min))}
		case f.Max != nil && size > *f.Max:
			problems[f.Key] = []string{t.Get("Shared.AtMost", fmt.Sprint(*f.Max))}
		}
	}
	return body, problems
}

type getter[T any] func() T

func (g getter[T]) Get() T { return g() }
