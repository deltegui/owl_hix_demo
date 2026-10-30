// Package entryweb is what the entries feature shows, in the browser and on the
// server: the models of the JSON API, the list, the editor and the viewer. It
// builds on both sides (the server renders the island and the browser hydrates
// it), so it imports only Hix, the UI packages and the pure domain, never OWL.
//
// The valtruc tags are only read by the server.
package entryweb

import "owldemo/web/ui/crud"

// Line is one line of an entry, as the editor writes it. Debit and Credit are
// pointers because exactly one of them must be there.
type Line struct {
	Account  int      `json:"account" valtruc:"min=99, max=600, required"`
	Debit    *float64 `json:"debit,omitempty" valtruc:"min=0"`
	Credit   *float64 `json:"credit,omitempty" valtruc:"min=0"`
	Currency string   `json:"currency" valtruc:"min=3, max=3, required"`
	Rate     float64  `json:"rate,omitempty"`    // read only: to the system currency
	SAmount  float64  `json:"samount,omitempty"` // read only: in the system currency
	Notes    string   `json:"notes,omitempty" valtruc:"max=200"`
}

// Entry is what the editor reads and writes. In basic mode an entry is an
// operation, an amount and two accounts; in advanced mode it is its lines.
type Entry struct {
	ID           int    `json:"id"`
	Mode         string `json:"mode"` // "basic" or "advanced"
	CanEditBasic bool   `json:"canEditBasic"`

	Concept    string `json:"concept" valtruc:"max=500"`
	CategoryId int    `json:"categoryId" valtruc:"required"`
	TagIds     []int  `json:"tagIds"`

	Date         string `json:"date" valtruc:"required"` // 2026-03-01
	Hour         int    `json:"hour" valtruc:"min=0, max=23"`
	Minute       int    `json:"minute" valtruc:"min=0, max=59"`
	ExchangeDate string `json:"exchangeDate"` // the day of the rates the lines use

	// Basic mode.
	OperationId   int      `json:"operationId,omitempty"`
	Amount        *float64 `json:"amount,omitempty" valtruc:"min=0"`
	Currency      string   `json:"currency,omitempty"`
	DebitAccount  int      `json:"debitAccount,omitempty"`
	CreditAccount int      `json:"creditAccount,omitempty"`

	// Advanced mode.
	Lines []Line `json:"lines,omitempty"`

	// Read only.
	CreatedAt string `json:"createdAt,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	UpdatedBy string `json:"updatedBy,omitempty"`
}

const (
	ModeBasic    = "basic"
	ModeAdvanced = "advanced"
)

// Row is a line of the list.
type Row struct {
	ID           int     `json:"id"`
	CarriedOutAt string  `json:"carriedOutAt"` // 2026-03-01 14:30
	Operation    string  `json:"operation"`    // key of the translation, empty for an entry that is not a basic operation
	SAmount      float64 `json:"samount"`
	Amount       float64 `json:"amount"`
	Sign         string  `json:"sign"` // of the currency of Amount
	Category     string  `json:"category"`
	Tags         string  `json:"tags"`
	CreatedBy    string  `json:"createdBy"`
}

// Listing is what the list endpoint answers: a page, and the total of the whole filter.
type Listing struct {
	Items []Row   `json:"items"`
	Page  int     `json:"page"`
	Pages int     `json:"pages"`
	Total int     `json:"total"`
	Sum   float64 `json:"sum"` // of every entry that matches, in the system currency
}

// Query is what the list depends on. Changing it loads another page. The
// strings are the values of the filter as the user typed them.
type Query struct {
	Concept, Category, Operation, From, To string
	OrderBy, Order                         string // Order is "asc" or "desc"
	Page                                   int
}

// Named is an item of a lookup list: a category or a tag.
type Named struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// AccountInfo is an account for the select lists. Name is the key of its translation.
type AccountInfo struct {
	Code  int    `json:"code"`
	Group int    `json:"group"`
	Name  string `json:"name"`
}

// OperationInfo is a basic operation for the select lists.
type OperationInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DebitGroup  int    `json:"debitGroup"`
	CreditGroup int    `json:"creditGroup"`
}

// Lookups is everything the screens pick from.
type Lookups struct {
	Categories   []Named         `json:"categories"`
	Tags         []Named         `json:"tags"`
	Accounts     []AccountInfo   `json:"accounts"`
	Operations   []OperationInfo `json:"operations"`
	Currencies   []string        `json:"currencies"`
	ExchangeDate string          `json:"exchangeDate"`
	// Advanced is the mode the editor opens a new entry in (a setting of the user).
	Advanced bool `json:"advanced"`
}

// State is the state of the island: the first page, the lookups and the
// translations. Both sides build the screen from it.
type State struct {
	Listing *Listing   `json:"listing"`
	Lookups Lookups    `json:"lookups"`
	Texts   crud.Texts `json:"t"`
}

// LineRequest is what the server needs to price a line: the line, and the day of the rates.
type LineRequest struct {
	Line         Line   `json:"line"`
	ExchangeDate string `json:"exchangeDate"`
}

// ConvertRequest asks the server to turn an entry into the other mode.
type ConvertRequest struct {
	To    string `json:"to"`
	Entry Entry  `json:"entry"`
}
