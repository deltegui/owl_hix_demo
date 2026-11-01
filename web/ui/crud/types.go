package crud

import "github.com/deltegui/owl/v3/island/client"

// Page is what a list endpoint answers.
type Page[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}

// Errors is the body of a 422 or 400 answer (see owl/island/client).
type Errors = client.Errors

// State is what the server sends with the island: the first page, and the
// language. Both sides build the screen from it.
type State[T any] struct {
	Page *Page[T] `json:"page"`
	Lang string   `json:"l"`
}
