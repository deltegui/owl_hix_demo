package entry

import (
	"strconv"
	"time"

	"github.com/deltegui/owl/v3/core"

	"owldemo/feature/entry/web"
	"owldemo/sys/currency"
	"owldemo/sys/ledger"
)

const (
	dateLayout     = "2006-01-02"
	dateTimeLayout = "2006-01-02 15:04"
)

// The rules of the domain that are not about a field. The code is the key of
// the translation in common/errors.json.
var (
	errUnbalanced = core.DomainError{Code: 2001, Message: "debit and credit must be the same amount"}
	errNotBasic   = core.DomainError{Code: 2002, Message: "the lines are not a basic operation"}
	errBadAccount = core.DomainError{Code: 2003, Message: "the accounts do not fit the operation"}
	errInvalid    = core.DomainError{Code: 2004, Message: "the entry is not valid"}
	errNotFound   = core.DomainError{Code: 2005, Message: "the entry does not exist"}
)

func euros(cents int64) float64 { return currency.FromCents(cents) }

func floatPtr(v float64) *float64 { return &v }

// toDTO is the entry for the editor. It opens in basic mode when its lines are
// a basic operation, and in advanced mode when they are not.
func toDTO(e ledger.Entry) entryweb.Entry {
	out := entryweb.Entry{
		ID: e.ID, Concept: e.Concept, CategoryId: e.CategoryID, TagIds: e.TagIDs,
		Date: e.CarriedOutAt.Format(dateLayout), Hour: e.CarriedOutAt.Hour(), Minute: e.CarriedOutAt.Minute(),
		ExchangeDate: e.ExchangeDate.Format(dateLayout),
		CreatedAt:    e.CreatedAt.Format(dateTimeLayout), CreatedBy: e.CreatedBy,
		UpdatedAt: e.UpdatedAt.Format(dateTimeLayout), UpdatedBy: e.UpdatedBy,
	}
	if basic, ok := ledger.ExtractBasic(e.Lines); ok {
		out.CanEditBasic = true
		out.Mode = entryweb.ModeBasic
		fillBasic(&out, basic)
		return out
	}
	out.Mode = entryweb.ModeAdvanced
	out.Lines = linesToDTO(e.Lines)
	return out
}

func fillBasic(out *entryweb.Entry, basic ledger.BasicEntry) {
	out.OperationId = basic.OperationID
	out.Amount = floatPtr(euros(basic.Amount))
	out.Currency = string(basic.Currency)
	out.DebitAccount = int(basic.Debit)
	out.CreditAccount = int(basic.Credit)
}

func linesToDTO(lines []ledger.Line) []entryweb.Line {
	out := make([]entryweb.Line, 0, len(lines))
	for _, l := range lines {
		line := entryweb.Line{Account: int(l.Account), Currency: string(l.Currency), Rate: l.Rate, SAmount: euros(l.SAmount), Notes: l.Notes}
		if l.Side == ledger.SideDebit {
			line.Debit = floatPtr(euros(l.Amount))
		} else {
			line.Credit = floatPtr(euros(l.Amount))
		}
		out = append(out, line)
	}
	return out
}

// toLines reads the lines of an advanced entry: a line is a debit or a credit.
func toLines(from []entryweb.Line) []ledger.Line {
	lines := make([]ledger.Line, 0, len(from))
	for _, l := range from {
		line := ledger.Line{Account: ledger.Account(l.Account), Currency: currency.Currency(l.Currency), Notes: l.Notes}
		if l.Debit != nil {
			line.Side, line.Amount = ledger.SideDebit, currency.ToCents(*l.Debit)
		} else if l.Credit != nil {
			line.Side, line.Amount = ledger.SideCredit, currency.ToCents(*l.Credit)
		}
		lines = append(lines, line)
	}
	return lines
}

// toDomain builds the entry to save from what the editor sent. In basic mode the
// lines are made from the operation; in advanced mode the operation is found
// from the lines, if they are one. It does the exchange to the system currency
// with the rates of the day of the entry and checks the entry balances.
func toDomain(dto entryweb.Entry) (ledger.Entry, error) {
	e := ledger.Entry{ID: dto.ID, Concept: dto.Concept, CategoryID: dto.CategoryId, TagIDs: dto.TagIds}

	day, err := time.Parse(dateLayout, dto.Date)
	if err != nil {
		return e, errInvalid.Wrap(err)
	}
	e.CarriedOutAt = time.Date(day.Year(), day.Month(), day.Day(), dto.Hour, dto.Minute, 0, 0, time.UTC)
	e.ExchangeDate = currency.LastExchangeDate()
	if rates, err := time.Parse(dateLayout, dto.ExchangeDate); err == nil {
		e.ExchangeDate = rates
	}

	if dto.Mode == entryweb.ModeBasic {
		op, ok := ledger.FindOperation(dto.OperationId)
		if !ok || dto.Amount == nil {
			return e, errInvalid
		}
		basic := ledger.BasicEntry{
			OperationID: op.ID, Amount: currency.ToCents(*dto.Amount), Currency: currency.Currency(dto.Currency),
			Debit: ledger.Account(dto.DebitAccount), Credit: ledger.Account(dto.CreditAccount),
		}
		if basic.Debit.Group() != op.DebitGroup || basic.Credit.Group() != op.CreditGroup {
			return e, errBadAccount.Wrapf("operation %q takes groups %d and %d, got %d and %d",
				op.Name, op.DebitGroup, op.CreditGroup, basic.Debit.Group(), basic.Credit.Group())
		}
		e.Lines = basic.ToLines()
	} else {
		e.Lines = toLines(dto.Lines)
	}

	e.FillSystemAmounts()
	if err := e.Validate(); err != nil {
		if err == ledger.ErrDebitCreditMustBeEqual {
			return e, errUnbalanced.Wrap(err)
		}
		return e, errInvalid.Wrap(err)
	}
	e.SAmount, _ = e.Totals()
	if basic, ok := ledger.ExtractBasic(e.Lines); ok {
		e.OperationID, e.Amount, e.BasicCurrency = basic.OperationID, basic.Amount, basic.Currency
	}
	return e, nil
}

// convert turns the entry of the editor into the other mode. Basic to advanced
// always works (a basic entry is two lines); advanced to basic only if the lines
// are a basic operation.
func convert(to string, dto entryweb.Entry) (entryweb.Entry, error) {
	if to == entryweb.ModeAdvanced {
		basic, err := toDomain(withMode(dto, entryweb.ModeBasic))
		if err != nil {
			// Not complete yet: the lines come out with what there is.
			op, ok := ledger.FindOperation(dto.OperationId)
			if !ok {
				return withMode(dto, entryweb.ModeAdvanced), nil
			}
			amount := 0.0
			if dto.Amount != nil {
				amount = *dto.Amount
			}
			lines := ledger.BasicEntry{
				OperationID: op.ID, Amount: currency.ToCents(amount), Currency: currency.Currency(dto.Currency),
				Debit: ledger.Account(dto.DebitAccount), Credit: ledger.Account(dto.CreditAccount),
			}.ToLines()
			out := withMode(dto, entryweb.ModeAdvanced)
			out.CanEditBasic = true
			out.Lines = priced(lines, dto.ExchangeDate)
			return out, nil
		}
		out := withMode(dto, entryweb.ModeAdvanced)
		out.CanEditBasic = true
		out.Lines = linesToDTO(basic.Lines)
		return out, nil
	}
	lines := toLines(dto.Lines)
	basic, ok := ledger.ExtractBasic(lines)
	if !ok {
		return dto, errNotBasic
	}
	out := withMode(dto, entryweb.ModeBasic)
	out.CanEditBasic = true
	out.Lines = nil
	fillBasic(&out, basic)
	return out, nil
}

func priced(lines []ledger.Line, exchangeDate string) []entryweb.Line {
	e := ledger.Entry{Lines: lines, ExchangeDate: currency.LastExchangeDate()}
	if day, err := time.Parse(dateLayout, exchangeDate); err == nil {
		e.ExchangeDate = day
	}
	e.FillSystemAmounts()
	return linesToDTO(e.Lines)
}

func withMode(dto entryweb.Entry, mode string) entryweb.Entry {
	dto.Mode = mode
	return dto
}

// blank is a new entry: today, in basic mode, as an income into the bank.
func blank(now time.Time) entryweb.Entry {
	return entryweb.Entry{
		Mode: entryweb.ModeBasic, CanEditBasic: true,
		Date: now.Format(dateLayout), Hour: now.Hour(), Minute: now.Minute(),
		ExchangeDate: currency.LastExchangeDate().Format(dateLayout),
		OperationId:  1, Currency: string(currency.System),
		DebitAccount: int(ledger.AssetBank), CreditAccount: int(ledger.IncomeSalary),
	}
}

// row is a line of the list.
func row(e ledger.Entry) entryweb.Row {
	r := entryweb.Row{
		ID: e.ID, CarriedOutAt: e.CarriedOutAt.Format(dateTimeLayout), SAmount: euros(e.SAmount),
		Category: categoryName(e.CategoryID), Tags: names(tags, e.TagIDs), CreatedBy: e.CreatedBy,
	}
	if op, ok := ledger.FindOperation(e.OperationID); ok {
		r.Operation = op.Name
		r.Amount, r.Sign = euros(e.Amount), e.BasicCurrency.Sign()
	} else {
		r.Amount, r.Sign = euros(e.SAmount), currency.System.Sign()
	}
	return r
}

func idString(id int) string { return strconv.Itoa(id) }
