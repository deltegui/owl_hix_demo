package entry

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/deltegui/owl/v3/pagination"

	"owldemo/feature/entry/web"
	"owldemo/sys/currency"
	"owldemo/sys/ledger"
	"owldemo/sys/persistence"
)

const pageSize = 8

// entryFilter is what the list asks for. Zero values ask for nothing.
type entryFilter struct {
	Concept    string
	CategoryID int
	Operation  int
	From, To   time.Time // To is the last day, included
}

func (f entryFilter) match(e ledger.Entry) bool {
	day := time.Date(e.CarriedOutAt.Year(), e.CarriedOutAt.Month(), e.CarriedOutAt.Day(), 0, 0, 0, 0, time.UTC)
	return (f.Concept == "" || strings.Contains(strings.ToLower(e.Concept), strings.ToLower(f.Concept))) &&
		(f.CategoryID == 0 || e.CategoryID == f.CategoryID) &&
		(f.Operation == 0 || e.OperationID == f.Operation) &&
		(f.From.IsZero() || !day.Before(f.From)) &&
		(f.To.IsZero() || !day.After(f.To))
}

// entryOrder is how the list is sorted: by one of the columns, up or down.
type entryOrder struct {
	By   string
	Desc bool
}

// less compares two entries by the column. The id breaks ties, so the order is the same on every call.
func (o entryOrder) compare(a, b ledger.Entry) int {
	var c int
	switch o.By {
	case "id":
		c = a.ID - b.ID
	case "samount":
		c = int(a.SAmount - b.SAmount)
	case "operation":
		c = strings.Compare(operationName(a), operationName(b))
	case "category":
		c = strings.Compare(categoryName(a.CategoryID), categoryName(b.CategoryID))
	default:
		c = a.CarriedOutAt.Compare(b.CarriedOutAt)
	}
	if c == 0 {
		c = a.ID - b.ID
	}
	if o.Desc {
		return -c
	}
	return c
}

func operationName(e ledger.Entry) string {
	op, _ := ledger.FindOperation(e.OperationID)
	return op.Name
}

type entryService struct {
	repo persistence.Repository[ledger.Entry]
	user string
}

func newEntryService() entryService {
	repo := persistence.NewMemory(
		func(e ledger.Entry) int { return e.ID },
		func(e ledger.Entry, id int) ledger.Entry { e.ID = id; return e },
		seed()...)
	return entryService{repo: repo, user: "demo"}
}

// Find returns a page of the entries that match, in that order, and the total
// of the system amounts of all of them (not only of the page).
func (service entryService) Find(filter entryFilter, order entryOrder, p pagination.Pagination) (pagination.List[ledger.Entry], int64, error) {
	all, err := service.repo.All(filter.match)
	if err != nil {
		return pagination.List[ledger.Entry]{}, 0, err
	}
	slices.SortFunc(all, order.compare)

	var sum int64
	for _, e := range all {
		sum += e.SAmount
	}
	pages := max(1, (len(all)+pageSize-1)/pageSize)
	current := min(max(p.CurrentPage, 1), pages)
	from, to := (current-1)*pageSize, min(current*pageSize, len(all))
	return pagination.List[ledger.Entry]{
		Items:      all[from:to],
		Pagination: pagination.Pagination{CurrentPage: current, ElementsPerPage: pageSize, TotalElements: len(all)},
	}, sum, nil
}

func (service entryService) FindOne(id int) (ledger.Entry, error) { return service.repo.FindOne(id) }

// Save creates the entry, or updates it if it has an id, and says who and when.
func (service entryService) Save(e ledger.Entry) (ledger.Entry, error) {
	now := time.Now().UTC()
	if e.ID == 0 {
		e.CreatedAt, e.CreatedBy = now, service.user
	} else {
		old, err := service.repo.FindOne(e.ID)
		if err != nil {
			return e, err
		}
		e.CreatedAt, e.CreatedBy = old.CreatedAt, old.CreatedBy
	}
	e.UpdatedAt, e.UpdatedBy = now, service.user
	return service.repo.Save(e)
}

func (service entryService) Delete(id int) error { return service.repo.Delete(id) }

// PriceLine says how many euros a unit of the currency of the line is worth on
// the day of the entry.
func (service entryService) PriceLine(l entryweb.Line, exchangeDate string) (entryweb.Line, error) {
	cur := currency.Currency(l.Currency)
	if !cur.Valid() {
		return l, fmt.Errorf("unknown currency %q", l.Currency)
	}
	day := currency.LastExchangeDate()
	if parsed, err := time.Parse(dateLayout, exchangeDate); err == nil {
		day = parsed
	}
	l.Rate = currency.Rate(cur, day)
	return l, nil
}

// seed is the entries the demo starts with: basic operations of every kind, in
// several currencies, and some that are not (more than two lines, or two
// currencies), which only the advanced mode can edit.
func seed() []ledger.Entry {
	type basic struct {
		op      int
		debit   ledger.Account
		credit  ledger.Account
		concept string
		cat     int
	}
	kinds := []basic{
		{1, ledger.AssetBank, ledger.IncomeSalary, "Salary", 3},
		{2, ledger.ExpenseOperational, ledger.AssetCash, "Groceries", 1},
		{2, ledger.ExpenseOperational, ledger.AssetBank, "Train ticket", 4},
		{3, ledger.AssetCash, ledger.AssetBank, "Cash withdrawal", 8},
		{2, ledger.ExpenseTaxes, ledger.AssetBank, "Property tax", 7},
		{4, ledger.AssetBank, ledger.LiabilityLoan, "Car loan", 4},
		{5, ledger.LiabilityLoan, ledger.AssetBank, "Loan payment", 4},
		{2, ledger.ExpenseOperational, ledger.AssetBank, "Dinner out", 5},
		{7, ledger.AssetBank, ledger.EquityAdjustment, "Found money", 8},
		{1, ledger.AssetBank, ledger.IncomeDividend, "Dividends", 8},
	}
	cur := []currency.Currency{currency.EUR, currency.EUR, currency.EUR, currency.USD, currency.EUR, currency.GBP}

	var out []ledger.Entry
	start := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	stamp := func(i int) (time.Time, time.Time) {
		day := start.AddDate(0, 0, i*3)
		return day.Add(time.Duration(i*7%24)*time.Hour + time.Duration(i*13%60)*time.Minute), day
	}
	for i := 0; i < 36; i++ {
		k := kinds[i%len(kinds)]
		when, day := stamp(i)
		b := ledger.BasicEntry{
			OperationID: k.op, Amount: currency.ToCents(float64(12+i*9) + 0.5), Currency: cur[i%len(cur)], Debit: k.debit, Credit: k.credit,
		}
		out = append(out, build(when, day, k.concept, k.cat, []int{1 + i%3}, b.ToLines()))
	}
	// Three lines: a salary with its taxes taken before it reaches the bank.
	for i := 0; i < 3; i++ {
		when, day := stamp(36 + i)
		out = append(out, build(when, day, "Salary after taxes", 3, []int{2, 5}, []ledger.Line{
			{Account: ledger.AssetBank, Side: ledger.SideDebit, Amount: currency.ToCents(1500 + float64(i)*50), Currency: currency.EUR},
			{Account: ledger.ExpenseTaxes, Side: ledger.SideDebit, Amount: currency.ToCents(400), Currency: currency.EUR},
			{Account: ledger.IncomeSalary, Side: ledger.SideCredit, Amount: currency.ToCents(1900 + float64(i)*50), Currency: currency.EUR, Notes: "gross"},
		}))
	}
	// Two currencies: dollars into the bank, paid with euros from cash.
	when, day := stamp(39)
	usd, _ := currency.ToSystem(25000, currency.USD, day)
	out = append(out, build(when, day, "Freelance job abroad", 3, []int{3}, []ledger.Line{
		{Account: ledger.AssetBank, Side: ledger.SideDebit, Amount: 25000, Currency: currency.USD},
		{Account: ledger.IncomeSales, Side: ledger.SideCredit, Amount: usd, Currency: currency.EUR},
	}))
	return out
}

func build(when, day time.Time, concept string, category int, tagIDs []int, lines []ledger.Line) ledger.Entry {
	e := ledger.Entry{
		Concept: concept, CategoryID: category, TagIDs: tagIDs, CarriedOutAt: when, ExchangeDate: day, Lines: lines,
		CreatedAt: when, CreatedBy: "demo", UpdatedAt: when, UpdatedBy: "demo",
	}
	e.FillSystemAmounts()
	e.SAmount, _ = e.Totals()
	if basic, ok := ledger.ExtractBasic(e.Lines); ok {
		e.OperationID, e.Amount, e.BasicCurrency = basic.OperationID, basic.Amount, basic.Currency
	}
	return e
}
