// Package ledger is the accounting of the demo, the same rules as LittleCash's
// sys/ledger: double entry (every entry balances debit and credit), accounts in
// groups, and the "basic operations" that let a user describe an entry of two
// lines as one sentence ("income of 100 from salary to bank").
//
// It is plain Go with no dependencies on the web, so the server uses it to
// decide and the browser to draw the same arrows and lists.
package ledger

import (
	"errors"
	"time"

	"owldemo/sys/currency"
)

// AccountGroup are the main groups of accounts.
type AccountGroup int

const (
	GroupAsset     AccountGroup = 1 // what you own or are owed
	GroupLiability AccountGroup = 2 // what you owe
	GroupEquity    AccountGroup = 3 // net worth
	GroupIncome    AccountGroup = 4 // raises net worth
	GroupExpense   AccountGroup = 5 // lowers net worth
)

type Side string

const (
	SideDebit  Side = "debit"
	SideCredit Side = "credit"
)

// Variation says if an amount on a side raises or lowers the balance of the account.
type Variation int

const (
	Increase Variation = 1
	Decrease Variation = 2
)

var debitVariation = map[AccountGroup]Variation{
	GroupAsset: Increase, GroupLiability: Decrease, GroupIncome: Decrease, GroupExpense: Increase, GroupEquity: Decrease,
}

var creditVariation = map[AccountGroup]Variation{
	GroupAsset: Decrease, GroupLiability: Increase, GroupIncome: Increase, GroupExpense: Decrease, GroupEquity: Increase,
}

func (g AccountGroup) Variation(side Side) Variation {
	if side == SideCredit {
		return creditVariation[g]
	}
	return debitVariation[g]
}

// Account is a code: the group is its hundreds.
type Account int

const (
	AssetBank       Account = 100
	AssetCash       Account = 101
	AssetInventory  Account = 102
	AssetFixed      Account = 103
	AssetReceivable Account = 104

	LiabilityLoan       Account = 200
	LiabilityObligation Account = 201

	EquityInitialCapital Account = 300
	EquityAdjustment     Account = 301

	IncomeSalary        Account = 400
	IncomeSales         Account = 401
	IncomeInterest      Account = 402
	IncomeDividend      Account = 403
	IncomeExtraordinary Account = 404

	ExpenseOperational Account = 500
	ExpenseTaxes       Account = 501
	ExpenseInterest    Account = 502
)

var accountNames = map[Account]string{
	AssetBank: "AccountAssetBank", AssetCash: "AccountAssetCash", AssetInventory: "AccountAssetInventory",
	AssetFixed: "AccountAssetFixed", AssetReceivable: "AccountAssetReceivable",
	LiabilityLoan: "AccountLiabilityLoan", LiabilityObligation: "AccountLiabilityObligation",
	EquityInitialCapital: "AccountEquityInitialCapital", EquityAdjustment: "AccountEquityAdjustment",
	IncomeSalary: "AccountIncomeSalary", IncomeSales: "AccountIncomeSales", IncomeInterest: "AccountIncomeInterest",
	IncomeDividend: "AccountIncomeDividend", IncomeExtraordinary: "AccountIncomeExtraordinary",
	ExpenseOperational: "AccountExpenseOperational", ExpenseTaxes: "AccountExpenseTaxes", ExpenseInterest: "AccountExpenseInterest",
}

// Accounts lists every account in order.
func Accounts() []Account {
	return []Account{
		AssetBank, AssetCash, AssetInventory, AssetFixed, AssetReceivable,
		LiabilityLoan, LiabilityObligation,
		EquityInitialCapital, EquityAdjustment,
		IncomeSalary, IncomeSales, IncomeInterest, IncomeDividend, IncomeExtraordinary,
		ExpenseOperational, ExpenseTaxes, ExpenseInterest,
	}
}

func (a Account) Group() AccountGroup { return AccountGroup(int(a) / 100) }

// Name is the key of the translation of the account.
func (a Account) Name() string { return accountNames[a] }

func (a Account) Valid() bool { _, ok := accountNames[a]; return ok }

// BasicOperation is a kind of entry of two balanced lines: which group of
// accounts goes in the debit and which in the credit.
type BasicOperation struct {
	ID          int
	Name        string // key of the translation
	DebitGroup  AccountGroup
	CreditGroup AccountGroup
	// ValueSide is the side to read to know how much the whole entry is worth.
	ValueSide Side
	// MustBeEqual: an entry only matches the operation if debit and credit are
	// the same amount.
	MustBeEqual bool
}

func (op BasicOperation) Group(side Side) AccountGroup {
	if side == SideCredit {
		return op.CreditGroup
	}
	return op.DebitGroup
}

// BasicOperations are the operations a user can pick in basic mode.
var BasicOperations = []BasicOperation{
	{1, "income", GroupAsset, GroupIncome, SideDebit, false},
	{2, "expense", GroupExpense, GroupAsset, SideCredit, false},
	{3, "transfer", GroupAsset, GroupAsset, SideDebit, true},
	{4, "loan received", GroupAsset, GroupLiability, SideDebit, true},
	{5, "loan payed", GroupLiability, GroupAsset, SideCredit, true},
	{6, "buy an asset", GroupAsset, GroupAsset, SideDebit, true},
	{7, "positive equity adjustment", GroupAsset, GroupEquity, SideDebit, true},
	{8, "negative equity adjustment", GroupEquity, GroupAsset, SideCredit, true},
}

func FindOperation(id int) (BasicOperation, bool) {
	for _, op := range BasicOperations {
		if op.ID == id {
			return op, true
		}
	}
	return BasicOperation{}, false
}

// Line is one side of an entry, for one account.
type Line struct {
	Account  Account
	Side     Side
	Amount   int64 // cents, in Currency
	Currency currency.Currency
	Rate     float64 // to the system currency, on the exchange date of the entry
	SAmount  int64   // cents, in the system currency
	Notes    string
}

// Entry is an accounting entry. Operation, Amount and BasicCurrency are only set
// when the lines are the two of a basic operation.
type Entry struct {
	ID           int
	Concept      string
	CategoryID   int
	TagIDs       []int
	CarriedOutAt time.Time
	ExchangeDate time.Time

	OperationID   int // 0 when the lines are not a basic operation
	Amount        int64
	SAmount       int64
	BasicCurrency currency.Currency

	Lines []Line

	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
}

var (
	ErrDebitCreditMustBeEqual = errors.New("ledger: debit and credit must be the same amount")
	ErrNoLines                = errors.New("ledger: an entry needs at least two lines")
	ErrInvalidLine            = errors.New("ledger: invalid line")
)

// FillSystemAmounts converts every line to the system currency with the rates of
// the exchange date.
func (e *Entry) FillSystemAmounts() {
	for i := range e.Lines {
		l := &e.Lines[i]
		l.SAmount, l.Rate = currency.ToSystem(l.Amount, l.Currency, e.ExchangeDate)
	}
}

// Totals are the sum of the debit side and of the credit side, in the system currency.
func (e Entry) Totals() (debit, credit int64) {
	for _, l := range e.Lines {
		if l.Side == SideDebit {
			debit += l.SAmount
		} else {
			credit += l.SAmount
		}
	}
	return
}

// Validate checks the lines are what an entry needs, and balanced.
func (e Entry) Validate() error {
	if len(e.Lines) < 2 {
		return ErrNoLines
	}
	for _, l := range e.Lines {
		if !l.Account.Valid() || !l.Currency.Valid() || l.Amount <= 0 {
			return ErrInvalidLine
		}
	}
	if debit, credit := e.Totals(); debit != credit {
		return ErrDebitCreditMustBeEqual
	}
	return nil
}

// BasicEntry is the entry as a basic user writes it: an operation, an amount
// and the two accounts.
type BasicEntry struct {
	OperationID int
	Amount      int64
	Currency    currency.Currency
	Debit       Account
	Credit      Account
}

// ToLines makes the two lines of a basic entry.
func (b BasicEntry) ToLines() []Line {
	return []Line{
		{Account: b.Debit, Side: SideDebit, Amount: b.Amount, Currency: b.Currency},
		{Account: b.Credit, Side: SideCredit, Amount: b.Amount, Currency: b.Currency},
	}
}

// ExtractBasic finds the basic operation the lines of an entry are, if they are
// one: a debit line and a credit line, in the same currency, whose accounts are in
// the groups of the operation (and with the same amount if it asks for that).
func ExtractBasic(lines []Line) (BasicEntry, bool) {
	if len(lines) != 2 {
		return BasicEntry{}, false
	}
	var debit, credit *Line
	for i := range lines {
		switch lines[i].Side {
		case SideDebit:
			debit = &lines[i]
		case SideCredit:
			credit = &lines[i]
		}
	}
	if debit == nil || credit == nil || debit.Currency != credit.Currency {
		return BasicEntry{}, false
	}
	for _, op := range BasicOperations {
		if debit.Account.Group() != op.DebitGroup || credit.Account.Group() != op.CreditGroup {
			continue
		}
		if op.MustBeEqual && debit.Amount != credit.Amount {
			continue
		}
		amount := debit.Amount
		if op.ValueSide == SideCredit {
			amount = credit.Amount
		}
		return BasicEntry{op.ID, amount, debit.Currency, debit.Account, credit.Account}, true
	}
	return BasicEntry{}, false
}
