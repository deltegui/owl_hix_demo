package forms

import (
	"reflect"
	"strconv"
	"strings"

	contactweb "owldemo/feature/contact/web"
	formsweb "owldemo/feature/forms/web"
	productweb "owldemo/feature/product/web"
)

// forms are described from the same structs the API validates, so the rules the
// browser sees are the rules the server enforces. A new form is a line here. (This
// is the one place where a feature knows others: it needs their models.)
var forms = map[string]formsweb.Form{
	"product": describe(productweb.Product{}, "Form.product", "/api/products", map[string][]string{"category": productweb.Categories}),
	"contact": describe(contactweb.Contact{}, "Form.contact", "/api/contacts", nil),
}

type formsService struct{}

func (formsService) find(name string) (formsweb.Form, bool) {
	form, ok := forms[name]
	return form, ok
}

// describe reads a struct: the json tag is the key, the valtruc tag the rules and
// the Go type the kind of control. A field with options is a select. The label is
// the key of the field's text: its name in Go. Fields without a json name, or
// named "id", are left out.
func describe(v any, title, endpoint string, options map[string][]string) formsweb.Form {
	t := reflect.TypeOf(v)
	form := formsweb.Form{Title: title, Endpoint: endpoint}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		key, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if key == "" || key == "-" || key == "id" {
			continue
		}
		field := formsweb.Field{Key: key, Label: f.Name, Options: options[key]}
		switch {
		case field.Options != nil:
			field.Kind = formsweb.Select
		case f.Type.Kind() == reflect.Bool:
			field.Kind = formsweb.Check
		case f.Type.Kind() >= reflect.Int && f.Type.Kind() <= reflect.Float64:
			field.Kind = formsweb.Number
		default:
			field.Kind = formsweb.Text
		}
		rules(&field, f.Tag.Get("valtruc"))
		form.Fields = append(form.Fields, field)
	}
	return form
}

// rules reads `min=3, max=60, required`. Other rules are left to the server.
func rules(field *formsweb.Field, tag string) {
	for _, rule := range strings.Split(tag, ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(rule), "=")
		n, err := strconv.ParseFloat(value, 64)
		switch {
		case name == "required":
			field.Required = true
		case name == "min" && err == nil:
			field.Min = &n
		case name == "max" && err == nil:
			field.Max = &n
		}
	}
}
