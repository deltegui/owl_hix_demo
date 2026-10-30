// Package formsweb draws any form from its description. The server builds the
// description from a struct and sends it with the page; the browser draws it
// without knowing the struct. It builds on both sides.
package formsweb

import "owldemo/web/ui/crud"

// Kinds of control.
const (
	Text   = "text"
	Number = "number"
	Select = "select"
	Check  = "check"
)

// Form is what a form screen needs: the key of its title, where to send it and
// the fields.
type Form struct {
	Title    string  `json:"title"`
	Endpoint string  `json:"endpoint"`
	Fields   []Field `json:"fields"`
}

// Field is one control. Label is the key of its text. Required, Min and Max
// come from the valtruc tags, so the browser can answer before asking the
// server (the server still decides).
type Field struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Kind     string   `json:"kind"`
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required,omitempty"`
	Min      *float64 `json:"min,omitempty"` // a value for Number, a length for Text
	Max      *float64 `json:"max,omitempty"`
}

// State is the state of the island: the form and the translations.
type State struct {
	Form  Form       `json:"form"`
	Texts crud.Texts `json:"t"`
}
