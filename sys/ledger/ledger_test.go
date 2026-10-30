package ledger

import (
	"testing"
	"time"

	"owldemo/sys/currency"
)

func TestVariation(t *testing.T) {
	// Money in the bank (asset) goes up on the debit side and down on the credit side.
	if AssetBank.Group().Variation(SideDebit) != Increase || AssetBank.Group().Variation(SideCredit) != Decrease {
		t.Error("asset variations are wrong")
	}
	if IncomeSalary.Group().Variation(SideCredit) != Increase || ExpenseOperational.Group().Variation(SideDebit) != Increase {
		t.Error("income and expense variations are wrong")
	}
}

func TestBasicEntryRoundTrip(t *testing.T) {
	basic := BasicEntry{OperationID: 1, Amount: 12345, Currency: currency.USD, Debit: AssetBank, Credit: IncomeSalary}
	lines := basic.ToLines()
	got, ok := ExtractBasic(lines)
	if !ok || got != basic {
		t.Fatalf("ExtractBasic = %+v, %v; want %+v", got, ok, basic)
	}
}

func TestExtractBasicRefuses(t *testing.T) {
	eur := currency.EUR
	cases := map[string][]Line{
		"three lines": {
			{Account: AssetBank, Side: SideDebit, Amount: 100, Currency: eur},
			{Account: ExpenseTaxes, Side: SideDebit, Amount: 50, Currency: eur},
			{Account: IncomeSalary, Side: SideCredit, Amount: 150, Currency: eur},
		},
		"two currencies": {
			{Account: AssetBank, Side: SideDebit, Amount: 100, Currency: eur},
			{Account: IncomeSalary, Side: SideCredit, Amount: 100, Currency: currency.USD},
		},
		"two debits": {
			{Account: AssetBank, Side: SideDebit, Amount: 100, Currency: eur},
			{Account: AssetCash, Side: SideDebit, Amount: 100, Currency: eur},
		},
		"a transfer with different amounts": {
			{Account: AssetBank, Side: SideDebit, Amount: 100, Currency: eur},
			{Account: AssetCash, Side: SideCredit, Amount: 90, Currency: eur},
		},
	}
	for name, lines := range cases {
		if _, ok := ExtractBasic(lines); ok {
			t.Errorf("%s: should not be a basic operation", name)
		}
	}
}

func TestExpenseReadsTheCreditSide(t *testing.T) {
	lines := []Line{
		{Account: ExpenseOperational, Side: SideDebit, Amount: 700, Currency: currency.EUR},
		{Account: AssetCash, Side: SideCredit, Amount: 700, Currency: currency.EUR},
	}
	got, ok := ExtractBasic(lines)
	if !ok || got.OperationID != 2 || got.Amount != 700 {
		t.Errorf("got %+v, %v", got, ok)
	}
}

func TestValidateBalance(t *testing.T) {
	date := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	entry := Entry{ExchangeDate: date, Lines: BasicEntry{OperationID: 1, Amount: 1000, Currency: currency.USD, Debit: AssetBank, Credit: IncomeSalary}.ToLines()}
	entry.FillSystemAmounts()
	if err := entry.Validate(); err != nil {
		t.Fatalf("a basic entry balances: %v", err)
	}
	// Two lines in different currencies with the same amount do not balance in euros.
	entry.Lines[1].Currency = currency.GBP
	entry.FillSystemAmounts()
	if err := entry.Validate(); err != ErrDebitCreditMustBeEqual {
		t.Errorf("Validate = %v, want %v", err, ErrDebitCreditMustBeEqual)
	}
	if err := (Entry{}).Validate(); err != ErrNoLines {
		t.Errorf("an empty entry: %v", err)
	}
}

func TestRateDependsOnTheDate(t *testing.T) {
	a := currency.Rate(currency.USD, time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC))
	b := currency.Rate(currency.USD, time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC))
	if a == b || currency.Rate(currency.EUR, time.Now()) != 1 {
		t.Errorf("rates: %v %v", a, b)
	}
}
