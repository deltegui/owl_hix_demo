// Package contactweb is what the contact feature shows, on the server and in the
// browser. See productweb.
package contactweb

import (
	"encoding/json"
	"fmt"

	"github.com/deltegui/hx"

	"owldemo/web/ui/crud"
)

type Contact struct {
	ID     int    `json:"id"`
	Name   string `json:"name" valtruc:"min=3, max=60, required"`
	Email  string `json:"email" valtruc:"min=5, max=120, required, contains=@"`
	Phone  string `json:"phone" valtruc:"max=20"`
	Active bool   `json:"active"`
}

// Screen is the list of contacts: the same generic screen as products, with
// another configuration.
func Screen(state crud.State[Contact]) hx.INode {
	return crud.Screen(crud.Config[Contact]{
		Texts:   TextsFor(state.Lang),
		API:     "/api/contacts",
		Initial: state.Page,
		ID:      func(c Contact) int { return c.ID },
		Name:    func(c Contact) string { return c.Name },
		Active:  func(c Contact) bool { return c.Active },
		Columns: []crud.Column[Contact]{
			crud.TextCol("#", func(c Contact) string { return fmt.Sprint(c.ID) }),
			crud.TextCol("Name", func(c Contact) string { return c.Name }),
			crud.TextCol("Email", func(c Contact) string { return c.Email }),
			crud.TextCol("Phone", func(c Contact) string { return c.Phone }),
			crud.ActiveCol("Status", func(c Contact) bool { return c.Active }),
		},
		Filters: []crud.Field{
			{Key: "q", Label: "Query"},
			{Key: "status", Label: "Status", Kind: crud.Select, Options: []string{"active", "disabled"}},
		},
		Form: []crud.Field{
			{Key: "name", Label: "Name"},
			{Key: "email", Label: "Email"},
			{Key: "phone", Label: "Phone"},
			{Key: "active", Label: "Active", Kind: crud.Check},
		},
	})
}

// ListIsland builds the screen from the state the server sent (JSON).
func ListIsland(state string) hx.INode {
	var s crud.State[Contact]
	if state != "" {
		json.Unmarshal([]byte(state), &s)
	}
	return Screen(s)
}
