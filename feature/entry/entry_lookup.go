package entry

import (
	"github.com/deltegui/valtruc"

	"owldemo/feature/entry/web"
	"owldemo/sys/currency"
	"owldemo/sys/ledger"
)

// In LittleCash the categories and the tags of a ledger are features of their
// own, with their own maintenance. Here they are fixed lists, which is all the
// entries need from them.
var categories = []entryweb.Named{
	{ID: 1, Name: "Food"}, {ID: 2, Name: "Housing"}, {ID: 3, Name: "Salary"}, {ID: 4, Name: "Transport"},
	{ID: 5, Name: "Leisure"}, {ID: 6, Name: "Health"}, {ID: 7, Name: "Taxes"}, {ID: 8, Name: "Savings"},
}

var tags = []entryweb.Named{
	{ID: 1, Name: "essential"}, {ID: 2, Name: "monthly"}, {ID: 3, Name: "one-off"},
	{ID: 4, Name: "shared"}, {ID: 5, Name: "deductible"}, {ID: 6, Name: "holiday"},
}

func names(list []entryweb.Named, ids []int) string {
	out := ""
	for _, id := range ids {
		for _, n := range list {
			if n.ID == id {
				if out != "" {
					out += ", "
				}
				out += n.Name
			}
		}
	}
	return out
}

func categoryName(id int) string { return names(categories, []int{id}) }

// lookups is everything the screens pick from.
func lookups() entryweb.Lookups {
	look := entryweb.Lookups{
		Categories:   categories,
		Tags:         tags,
		ExchangeDate: currency.LastExchangeDate().Format(dateLayout),
	}
	for _, acc := range ledger.Accounts() {
		look.Accounts = append(look.Accounts, entryweb.AccountInfo{Code: int(acc), Group: int(acc.Group()), Name: acc.Name()})
	}
	for _, op := range ledger.BasicOperations {
		look.Operations = append(look.Operations, entryweb.OperationInfo{
			ID: op.ID, Name: op.Name, DebitGroup: int(op.DebitGroup), CreditGroup: int(op.CreditGroup)})
	}
	for _, c := range currency.All() {
		look.Currencies = append(look.Currencies, string(c))
	}
	return look
}

// valtrucID is the identifier of a custom validation error, which is the key of
// its message in the file of the view.
func valtrucID(key string) valtruc.ValidatorIdentifier { return valtruc.ValidatorIdentifier(key) }
