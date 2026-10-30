package entry

import (
	"errors"
	"testing"

	"github.com/deltegui/owl/v3/core"

	"owldemo/feature/entry/web"
)

func basicDTO() entryweb.Entry {
	amount := 100.0
	return entryweb.Entry{
		Mode: entryweb.ModeBasic, Date: "2026-03-01", Hour: 10, Minute: 5, ExchangeDate: "2026-03-01", CategoryId: 3,
		OperationId: 1, Amount: &amount, Currency: "USD", DebitAccount: 100, CreditAccount: 400,
	}
}

func TestBasicToAdvancedAndBack(t *testing.T) {
	advanced, err := convert(entryweb.ModeAdvanced, basicDTO())
	if err != nil || advanced.Mode != entryweb.ModeAdvanced || len(advanced.Lines) != 2 {
		t.Fatalf("to advanced: %+v, %v", advanced, err)
	}
	if advanced.Lines[0].Debit == nil || *advanced.Lines[0].Debit != 100 || advanced.Lines[1].Credit == nil || advanced.Lines[0].Rate == 1 {
		t.Errorf("the lines are a debit and a credit of 100 dollars, priced: %+v", advanced.Lines)
	}
	back, err := convert(entryweb.ModeBasic, advanced)
	if err != nil || back.OperationId != 1 || *back.Amount != 100 || back.Currency != "USD" || back.Lines != nil {
		t.Errorf("back to basic: %+v, %v", back, err)
	}
}

func TestAdvancedThatIsNotBasicCannotBecomeBasic(t *testing.T) {
	three := entryweb.Entry{Mode: entryweb.ModeAdvanced, Lines: []entryweb.Line{
		{Account: 100, Debit: floatPtr(10), Currency: "EUR"},
		{Account: 501, Debit: floatPtr(5), Currency: "EUR"},
		{Account: 400, Credit: floatPtr(15), Currency: "EUR"},
	}}
	_, err := convert(entryweb.ModeBasic, three)
	var domain core.DomainError
	if !errors.As(err, &domain) || domain.Code != errNotBasic.Code {
		t.Errorf("err = %v", err)
	}
}

func TestToDomain(t *testing.T) {
	e, err := toDomain(basicDTO())
	if err != nil || e.OperationID != 1 || e.Amount != 10000 || len(e.Lines) != 2 || e.SAmount == 0 || e.CarriedOutAt.Minute() != 5 {
		t.Fatalf("toDomain = %+v, %v", e, err)
	}

	wrongGroup := basicDTO()
	wrongGroup.DebitAccount = 500 // an income puts money in an asset, not in an expense
	if _, err := toDomain(wrongGroup); !isCode(err, errBadAccount.Code) {
		t.Errorf("accounts that do not fit: %v", err)
	}

	unbalanced := entryweb.Entry{Mode: entryweb.ModeAdvanced, Date: "2026-03-01", ExchangeDate: "2026-03-01", Lines: []entryweb.Line{
		{Account: 100, Debit: floatPtr(10), Currency: "EUR"}, {Account: 400, Credit: floatPtr(11), Currency: "EUR"},
	}}
	if _, err := toDomain(unbalanced); !isCode(err, errUnbalanced.Code) {
		t.Errorf("unbalanced: %v", err)
	}

	badDate := basicDTO()
	badDate.Date = "tomorrow"
	if _, err := toDomain(badDate); !isCode(err, errInvalid.Code) {
		t.Errorf("a bad date: %v", err)
	}
}

func TestAdvancedFindsTheOperationOfItsLines(t *testing.T) {
	dto := entryweb.Entry{Mode: entryweb.ModeAdvanced, Date: "2026-03-01", ExchangeDate: "2026-03-01", Lines: []entryweb.Line{
		{Account: 500, Debit: floatPtr(7), Currency: "EUR"}, {Account: 101, Credit: floatPtr(7), Currency: "EUR"},
	}}
	e, err := toDomain(dto)
	if err != nil || e.OperationID != 2 || e.Amount != 700 {
		t.Errorf("an expense paid in cash: %+v, %v", e, err)
	}
	if out := toDTO(e); out.Mode != entryweb.ModeBasic || !out.CanEditBasic || out.OperationId != 2 {
		t.Errorf("it opens as a sentence: %+v", out)
	}
}

func isCode(err error, code int) bool {
	var domain core.DomainError
	return errors.As(err, &domain) && domain.Code == code
}
