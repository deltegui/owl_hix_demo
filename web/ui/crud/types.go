package crud

// Page is what a list endpoint answers.
type Page[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}

// Errors is the body of a 422 answer: what is wrong, by field. A rule of the
// domain that is not about a field answers with Message instead.
type Errors struct {
	Errors  map[string][]string `json:"errors,omitempty"`
	Message string              `json:"error,omitempty"`
}

// State is what the server sends with the island: the first page, and the
// language. Both sides build the screen from it.
type State[T any] struct {
	Page *Page[T] `json:"page"`
	Lang string   `json:"l"`
}
