package crud

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/deltegui/hx"
)

type Kind int

const (
	Text Kind = iota
	Number
	Select
	Check
)

// Field describes a filter or a form control. Key is the JSON name of the property.
type Field struct {
	Key     string
	Label   string // the key of its text (see Texts)
	Kind    Kind
	Options []string // for Select, the values; the text of each is "Opt.<value>"
}

// formState holds what the user is typing, one signal per field.
type formState struct {
	fields []Field
	text   map[string]*hx.SignalT[string]
	checks map[string]*hx.SignalT[bool]
	errors *hx.SignalT[map[string][]string]
	other  *hx.SignalT[string] // an error that is not about a field
}

func newFormState(fields []Field) *formState {
	f := &formState{
		fields: fields,
		text:   map[string]*hx.SignalT[string]{},
		checks: map[string]*hx.SignalT[bool]{},
		errors: hx.Signal(map[string][]string{}),
		other:  hx.Signal(""),
	}
	for _, field := range fields {
		if field.Kind == Check {
			f.checks[field.Key] = hx.Signal(false)
		} else {
			f.text[field.Key] = hx.Signal("")
		}
	}
	return f
}

// load fills the form with an item. A nil item is an empty form for a new one;
// checkboxes start checked, since a new thing is usually enabled.
func (f *formState) load(item any) {
	values := map[string]any{}
	if item != nil {
		data, _ := json.Marshal(item)
		json.Unmarshal(data, &values)
	}
	for _, field := range f.fields {
		raw, ok := values[field.Key]
		switch field.Kind {
		case Check:
			checked, _ := raw.(bool)
			f.checks[field.Key].Set(checked || !ok)
		case Number:
			if number, isNumber := raw.(float64); isNumber {
				f.text[field.Key].Set(strconv.FormatFloat(number, 'f', -1, 64))
			} else {
				f.text[field.Key].Set("")
			}
		default:
			f.text[field.Key].Set(fmt.Sprint(orEmpty(raw)))
		}
	}
	f.errors.Set(map[string][]string{})
	f.other.Set("")
}

func orEmpty(v any) any {
	if v == nil {
		return ""
	}
	return v
}

// body is what gets sent to the server, with each value as the type the JSON needs.
func (f *formState) body() map[string]any {
	body := map[string]any{}
	for _, field := range f.fields {
		switch field.Kind {
		case Check:
			body[field.Key] = f.checks[field.Key].Get()
		case Number:
			number, _ := strconv.ParseFloat(strings.TrimSpace(f.text[field.Key].Get()), 64)
			body[field.Key] = number
		default:
			body[field.Key] = strings.TrimSpace(f.text[field.Key].Get())
		}
	}
	return body
}

func (f *formState) errorsOf(key string) hx.Gettable[[]string] {
	return getter[[]string](func() []string { return f.errors.Get()[key] })
}

// getter turns a function into something that can be read, so it can be passed
// where a signal is expected. The reads inside it are tracked by whoever calls Get.
type getter[T any] func() T

func (g getter[T]) Get() T { return g() }
