package entryweb

import (
	"context"
	"strconv"
	"strings"

	"github.com/deltegui/hx"

	"owldemo/sys/currency"
	"owldemo/sys/ledger"
	"owldemo/web/ui/bs"
	"owldemo/web/ui/crud"
)

// lineState is what the user is typing in a line of an advanced entry.
type lineState struct {
	key      int
	account  *hx.SignalT[string]
	debit    *hx.SignalT[string]
	credit   *hx.SignalT[string]
	currency *hx.SignalT[string]
	notes    *hx.SignalT[string]
	rate     *hx.SignalT[float64] // to the system currency, from the server
	pricing  *hx.ActionT[string, Line]
	errs     *hx.SignalT[map[string][]string]
}

func (l *lineState) amount() (float64, bool) {
	if v, ok := parseFloat(hx.UntrackGet(l.debit)); ok {
		return v, true
	}
	return parseFloat(hx.UntrackGet(l.credit))
}

// editor is the form of an entry: basic mode is a sentence, advanced mode is a
// table of lines. Everything the user types is in signals; the server decides
// what is valid.
func editor(a *app, e Entry) hx.INode {
	t := a.t

	// What every entry has.
	concept := hx.Signal(e.Concept)
	category := hx.Signal(itoa(e.CategoryId))
	tags := hx.Signal(map[int]bool{})
	for _, id := range e.TagIds {
		hx.UntrackGet(tags)[id] = true
	}
	date := hx.Signal(e.Date)
	hour := hx.Signal(strconv.Itoa(e.Hour))
	minute := hx.Signal(strconv.Itoa(e.Minute))
	errs := hx.Signal(map[string][]string{})
	other := hx.Signal("")

	// Basic mode.
	mode := hx.Signal(e.Mode)
	canBasic := hx.Signal(e.CanEditBasic)
	operation := hx.Signal(itoa(e.OperationId))
	amount := hx.Signal("")
	basicCurrency := hx.Signal(orDefault(e.Currency, string(currency.System)))
	debitAccount := hx.Signal(itoa(e.DebitAccount))
	creditAccount := hx.Signal(itoa(e.CreditAccount))
	if e.Amount != nil {
		amount.Set(strconv.FormatFloat(*e.Amount, 'f', -1, 64))
	}

	// Advanced mode.
	nextKey := 0
	newLine := func(l Line) *lineState {
		nextKey++
		s := &lineState{
			key: nextKey, account: hx.Signal(itoa(l.Account)), debit: hx.Signal(""), credit: hx.Signal(""),
			currency: hx.Signal(orDefault(l.Currency, string(currency.System))), notes: hx.Signal(l.Notes),
			rate: hx.Signal(orRate(l.Rate)), errs: hx.Signal(map[string][]string{}),
		}
		// The rate comes from the server and depends on the currency and the day of
		// the entry; the amount in euros is worked out here. Each line has its own
		// action: the last call of one would cancel the call of another.
		s.pricing = hx.Action(func(ctx context.Context, cur string) (Line, error) {
			return priceLine(ctx, LineRequest{Line: Line{Account: parseInt(hx.UntrackGet(s.account)), Currency: cur}, ExchangeDate: e.ExchangeDate})
		})
		s.pricing.OnSuccess(func(priced Line, cur string) {
			if hx.UntrackGet(s.currency) == cur { // the user may have changed it again
				s.rate.Set(orRate(priced.Rate))
			}
		})
		if l.Debit != nil {
			s.debit.Set(strconv.FormatFloat(*l.Debit, 'f', -1, 64))
		}
		if l.Credit != nil {
			s.credit.Set(strconv.FormatFloat(*l.Credit, 'f', -1, 64))
		}
		return s
	}
	linesOf := func(from []Line) []*lineState {
		out := make([]*lineState, len(from))
		for i, l := range from {
			out[i] = newLine(l)
		}
		return out
	}
	lines := hx.Signal(linesOf(e.Lines))

	refreshRate := func(l *lineState) { l.pricing.Run(hx.UntrackGet(l.currency)) }

	// The sums, in the system currency, of the two sides.
	sum := func(side func(*lineState) *hx.SignalT[string]) *hx.ComputedT[int64] {
		return hx.Computed(func() int64 {
			var total int64
			for _, l := range lines.Get() {
				if v, ok := parseFloat(side(l).Get()); ok {
					total += cents(v, l.rate.Get())
				}
			}
			return total
		})
	}
	debitSum := sum(func(l *lineState) *hx.SignalT[string] { return l.debit })
	creditSum := sum(func(l *lineState) *hx.SignalT[string] { return l.credit })
	balanced := hx.Computed(func() bool { return debitSum.Get() == creditSum.Get() })

	// collect is the entry as the form has it.
	collect := func() Entry {
		out := Entry{
			ID: e.ID, Mode: hx.UntrackGet(mode), CanEditBasic: hx.UntrackGet(canBasic), ExchangeDate: e.ExchangeDate,
			Concept:    strings.TrimSpace(hx.UntrackGet(concept)),
			CategoryId: parseInt(hx.UntrackGet(category)),
			Date:       hx.UntrackGet(date), Hour: parseInt(hx.UntrackGet(hour)), Minute: parseInt(hx.UntrackGet(minute)),
		}
		for _, n := range a.look.Tags {
			if hx.UntrackGet(tags)[n.ID] {
				out.TagIds = append(out.TagIds, n.ID)
			}
		}
		if out.Mode == ModeBasic {
			out.OperationId = parseInt(hx.UntrackGet(operation))
			if v, ok := parseFloat(hx.UntrackGet(amount)); ok {
				out.Amount = &v
			}
			out.Currency = hx.UntrackGet(basicCurrency)
			out.DebitAccount = parseInt(hx.UntrackGet(debitAccount))
			out.CreditAccount = parseInt(hx.UntrackGet(creditAccount))
			return out
		}
		for _, l := range hx.UntrackGet(lines) {
			line := Line{Account: parseInt(hx.UntrackGet(l.account)), Currency: hx.UntrackGet(l.currency), Notes: hx.UntrackGet(l.notes)}
			if v, ok := parseFloat(hx.UntrackGet(l.debit)); ok {
				line.Debit = &v
			}
			if v, ok := parseFloat(hx.UntrackGet(l.credit)); ok {
				line.Credit = &v
			}
			out.Lines = append(out.Lines, line)
		}
		return out
	}

	// apply puts what the server answered in the form (after changing the mode).
	apply := func(n Entry) {
		mode.Set(n.Mode)
		canBasic.Set(n.CanEditBasic)
		operation.Set(itoa(n.OperationId))
		amount.Set("")
		if n.Amount != nil {
			amount.Set(strconv.FormatFloat(*n.Amount, 'f', -1, 64))
		}
		basicCurrency.Set(orDefault(n.Currency, string(currency.System)))
		debitAccount.Set(itoa(n.DebitAccount))
		creditAccount.Set(itoa(n.CreditAccount))
		lines.Set(linesOf(n.Lines))
		errs.Set(map[string][]string{})
	}

	// Saving.
	save := hx.Action(func(ctx context.Context, n Entry) (Entry, error) { return saveEntry(ctx, n) })
	save.OnSuccess(func(saved Entry, _ Entry) {
		a.close()
		a.notice.Set(t.Get("SavedMessage", saved.ID))
		a.reload()
	})
	save.OnError(func(err error, sent Entry) {
		if byField, ok := crud.FieldErrors(err); ok {
			errs.Set(byField)
			for i, l := range hx.UntrackGet(lines) {
				l.errs.Set(subErrors(byField, "lines["+strconv.Itoa(i)+"]."))
			}
			return
		}
		if message, ok := crud.ErrorMessage(err); ok {
			other.Set(message)
			return
		}
		other.Set(err.Error())
	})
	submit := func() {
		errs.Set(map[string][]string{})
		for _, l := range hx.UntrackGet(lines) {
			l.errs.Set(map[string][]string{})
		}
		other.Set("")
		n := collect()
		if n.Mode == ModeAdvanced && !hx.UntrackGet(balanced) {
			other.Set(t.Get("DebitCreditMustBeEqualMessage"))
			return
		}
		save.Run(n)
	}

	// Changing the mode is a decision of the server: a basic entry becomes two
	// lines, and some lines are not an operation.
	convert := hx.Action(func(ctx context.Context, r ConvertRequest) (Entry, error) { return convertEntry(ctx, r) })
	convert.OnSuccess(func(n Entry, _ ConvertRequest) { apply(n) })
	convert.OnError(func(err error, _ ConvertRequest) {
		if message, ok := crud.ErrorMessage(err); ok {
			other.Set(message)
			return
		}
		other.Set(err.Error())
	})
	toggleMode := func() {
		other.Set("")
		to := ModeAdvanced
		if hx.UntrackGet(mode) == ModeAdvanced {
			to = ModeBasic
		}
		convert.Run(ConvertRequest{To: to, Entry: collect()})
	}

	errsOf := func(key string) hx.Gettable[[]string] {
		return getter[[]string](func() []string { return errs.Get()[key] })
	}

	title := t.Get("NewTitle")
	if e.ID != 0 {
		title = t.Get("EditTitle", e.ID)
	}

	// The part that depends on the mode.
	operationBody := hx.Dynamic(func() hx.INode {
		basic := mode.Get() == ModeBasic // read here: inside Untrack it would not be followed
		var node hx.INode
		hx.Untrack(func() {
			if basic {
				node = basicPhrase(a, operation, amount, basicCurrency, creditAccount, debitAccount, errsOf)
			} else {
				node = advancedLines(a, lines, newLine, refreshRate, debitSum, creditSum, balanced)
			}
		})
		return node
	})

	modeButton := bs.Button("outline-secondary", "")
	modeButton.Class("mb-3").Attribute("data-role", "mode")
	modeButton.Effect(func() {
		if mode.Get() == ModeBasic {
			modeButton.Text(t.Get("Mode.GoAdvanced"))
		} else {
			modeButton.Text(t.Get("Mode.GoBasic"))
		}
		if mode.Get() == ModeAdvanced && !canBasic.Get() {
			modeButton.Attribute("disabled", "disabled")
		} else {
			modeButton.RemoveAttribute("disabled")
		}
	})
	modeButton.OnClick(func(hx.EventContext) { toggleMode() })

	hint := hx.P().Class("text-muted", "small")
	hint.Effect(func() {
		if mode.Get() == ModeAdvanced && !canBasic.Get() {
			hint.Text(t.Get("WhyBasicUserCantEditOperation"))
		} else {
			hint.Text("")
		}
	})

	categories := make([]bs.Option, 0, len(a.look.Categories))
	for _, c := range a.look.Categories {
		categories = append(categories, bs.Option{Value: strconv.Itoa(c.ID), Label: c.Name})
	}

	return hx.Div().Attribute("data-layer", "editor").Body(
		bs.Row(
			bs.Col(8, hx.H1().Text(title), hx.P().Class("text-muted").Text(t.Get("EditSubtitle"))),
			bs.Col(4, bs.Button("secondary", t.Get("Shared.Return")).Class("float-end").OnClick(func(hx.EventContext) { a.close() }))),
		bs.Alert("danger", other, func() { other.Set("") }),
		hx.Form().OnSubmit(func(ctx hx.EventContext) {
			bs.PreventDefault(ctx)
			submit()
		}).Body(
			bs.Card(t.Get("Operation"), modeButton, hint, operationBody),
			bs.Card(t.Get("CarriedOutAt"), bs.Row(
				bs.Col(4, bs.Field(t.Get("Day"), bs.DateInput(date), errsOf("date"))),
				bs.Col(2, bs.Field(t.Get("Hour"), bs.NumberInput(hour), errsOf("hour"))),
				bs.Col(2, bs.Field(t.Get("Minute"), bs.NumberInput(minute), errsOf("minute"))))),
			bs.Card(t.Get("Categories"), bs.Row(
				bs.Col(4, bs.Field(t.Get("Category"), bs.Select(category, t.Get("Shared.Choose"), categories).Attribute("data-field", "category"), errsOf("categoryId"))),
				bs.Col(8, hx.Div().Body(hx.Label().Class("form-label").Text(t.Get("Tags")), tagChips(a, tags))))),
			bs.Card(t.Get("Concept"), bs.Field("", bs.TextArea(concept), errsOf("concept"))),
			saveButton(t, save.Pending(), submit)))
}

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func orRate(r float64) float64 {
	if r == 0 {
		return 1
	}
	return r
}

// subErrors are the errors of a prefix, without it: "lines[1].debit" is "debit" in line 1.
func subErrors(all map[string][]string, prefix string) map[string][]string {
	out := map[string][]string{}
	for key, messages := range all {
		if strings.HasPrefix(key, prefix) {
			out[strings.TrimPrefix(key, prefix)] = messages
		}
	}
	return out
}

func saveButton(t crud.Texts, pending hx.Gettable[bool], submit func()) hx.INode {
	button := bs.SubmitButton("primary", t.Get("Shared.Save"))
	button.Class("float-end")
	button.Effect(func() {
		if pending.Get() {
			button.Text(t.Get("Shared.Saving")).Attribute("disabled", "disabled")
		} else {
			button.Text(t.Get("Shared.Save")).RemoveAttribute("disabled")
		}
	})
	return button
}

// tagChips are the tags as buttons that are on or off: a select of many values.
func tagChips(a *app, selected *hx.SignalT[map[int]bool]) hx.INode {
	chips := make([]hx.INode, 0, len(a.look.Tags))
	for _, tag := range a.look.Tags {
		tag := tag
		chip := hx.Button().Type("button").Class("btn", "btn-sm", "me-1", "mb-1").Text(tag.Name).Attribute("data-tag", strconv.Itoa(tag.ID))
		chip.Effect(func() {
			on := selected.Get()[tag.ID]
			chip.ClassIf("btn-primary", on).ClassIf("btn-outline-secondary", !on)
		})
		chips = append(chips, chip.OnClick(func(hx.EventContext) {
			next := map[int]bool{}
			for id, on := range hx.UntrackGet(selected) {
				next[id] = on
			}
			next[tag.ID] = !next[tag.ID]
			selected.Set(next)
		}))
	}
	return hx.Div().BodyList(chips)
}

// basicPhrase is the sentence of a basic entry: "Income of 100 EUR from Salary to Bank".
func basicPhrase(a *app, operation, amount, cur, credit, debit *hx.SignalT[string], errsOf func(string) hx.Gettable[[]string]) hx.INode {
	t := a.t
	operations := make([]bs.Option, 0, len(a.look.Operations))
	for _, op := range a.look.Operations {
		operations = append(operations, bs.Option{Value: strconv.Itoa(op.ID), Label: t.Get(op.Name)})
	}
	currencies := bs.Plain(a.look.Currencies...)

	// The accounts a side can have depend on the operation: an income takes
	// money from an income account into an asset one.
	accountOptions := func(side ledger.Side) []bs.Option {
		op, _ := a.operation(parseInt(hx.UntrackGet(operation)))
		group := op.DebitGroup
		if side == ledger.SideCredit {
			group = op.CreditGroup
		}
		var out []bs.Option
		for _, acc := range a.accountsOf(group) {
			out = append(out, bs.Option{Value: strconv.Itoa(acc.Code), Label: t.Get(acc.Name)})
		}
		return out
	}
	accountSelect := func(side ledger.Side, value *hx.SignalT[string]) hx.INode {
		return hx.Div().Class("d-inline-block").Body(hx.Dynamic(func() hx.INode {
			operation.Get() // the options follow the operation
			var node hx.INode
			hx.Untrack(func() { node = bs.Select(value, "", accountOptions(side)).Attribute("data-side", string(side)) })
			return node
		}))
	}
	// When the operation changes, an account that is not in the group any more
	// is replaced by the first one that is.
	fix := func(side ledger.Side, value *hx.SignalT[string]) {
		options := accountOptions(side)
		current := hx.UntrackGet(value)
		for _, o := range options {
			if o.Value == current {
				return
			}
		}
		if len(options) > 0 {
			value.Set(options[0].Value)
		} else {
			value.Set("")
		}
	}

	part := func(controls ...hx.INode) hx.INode {
		return hx.Div().Class("d-flex", "flex-wrap", "align-items-center", "gap-2", "mb-2").Attribute("data-role", "phrase").Body(controls...)
	}
	return hx.Div().Body(
		part(
			bs.Select(operation, t.Get("Shared.Choose"), operations).Class("w-auto").Attribute("data-field", "operation").OnChange(func(hx.EventContext) {
				fix(ledger.SideDebit, debit)
				fix(ledger.SideCredit, credit)
			}),
			hx.Span().Text(t.Get("Phrase.Of")),
			bs.NumberInput(amount).Class("w-auto").Attribute("data-field", "amount").Style("width", "9rem"),
			bs.Select(cur, "", currencies).Class("w-auto"),
			hx.Span().Text(t.Get("Phrase.From")),
			accountSelect(ledger.SideCredit, credit),
			hx.Span().Text(t.Get("Phrase.To")),
			accountSelect(ledger.SideDebit, debit)),
		hx.Div().Class("invalid-feedback", "d-block").Body(
			errorLine(errsOf("operationId")), errorLine(errsOf("amount")), errorLine(errsOf("creditAccount")), errorLine(errsOf("debitAccount"))))
}

// errorLine is the messages of a field as text.
func errorLine(errs hx.Gettable[[]string]) hx.INode {
	line := hx.Div()
	line.Effect(func() { line.Text(strings.Join(errs.Get(), " · ")) })
	return line
}

// advancedLines is the table of lines of an advanced entry, with the sums of
// both sides in red until they are the same.
func advancedLines(a *app, lines *hx.SignalT[[]*lineState], newLine func(Line) *lineState, refreshRate func(*lineState),
	debitSum, creditSum *hx.ComputedT[int64], balanced *hx.ComputedT[bool]) hx.INode {
	t := a.t

	accounts := make([]bs.Option, 0, len(a.look.Accounts))
	for _, acc := range a.look.Accounts {
		accounts = append(accounts, bs.Option{Value: strconv.Itoa(acc.Code), Label: t.Get(acc.Name)})
	}
	currencies := bs.Plain(a.look.Currencies...)

	// arrow says if an amount on that side raises or lowers the balance of the account.
	arrow := func(l *lineState, side ledger.Side) hx.INode {
		span := hx.Span().Class("text-muted", "ms-1")
		span.Effect(func() {
			acc, ok := a.account(parseInt(l.account.Get()))
			switch {
			case !ok:
				span.Text("")
			case ledger.AccountGroup(acc.Group).Variation(side) == ledger.Increase:
				span.Text("↑")
			default:
				span.Text("↓")
			}
		})
		return span
	}
	// A line has an amount on one side: while one has a value the other is hidden.
	amountCell := func(l *lineState, side ledger.Side, mine, theirs *hx.SignalT[string]) hx.INode {
		box := hx.Div().Attribute("data-side", string(side))
		box.Effect(func() { box.ClassIf("d-none", theirs.Get() != "" && mine.Get() == "") })
		input := bs.NumberInput(mine).Attribute("data-field", string(side))
		return box.Body(
			hx.Div().Class("d-flex", "align-items-center").Body(input, arrow(l, side)),
			errorLine(getter[[]string](func() []string { return l.errs.Get()[string(side)] })))
	}

	row := func(key int, item hx.Gettable[*lineState], _ hx.Gettable[int]) hx.INode {
		l := hx.UntrackGet(item)
		inEuros := hx.Computed(func() string {
			amount, ok := parseFloat(l.debit.Get())
			if !ok {
				amount, ok = parseFloat(l.credit.Get())
			}
			if !ok {
				return ""
			}
			return money(float64(cents(amount, l.rate.Get()))/100) + " €"
		})
		return hx.Tr().Attribute("data-line", strconv.Itoa(key)).Body(
			hx.Td().Body(
				bs.Select(l.account, "", accounts).Attribute("data-field", "account"),
				errorLine(getter[[]string](func() []string { return l.errs.Get()["account"] }))),
			hx.Td().Body(amountCell(l, ledger.SideDebit, l.debit, l.credit)),
			hx.Td().Body(amountCell(l, ledger.SideCredit, l.credit, l.debit)),
			hx.Td().Body(bs.Select(l.currency, "", currencies).Style("min-width", "6rem").OnChange(func(hx.EventContext) { refreshRate(l) }).Attribute("data-field", "currency")),
			hx.Td().Class("text-end", "text-nowrap").Body(hx.Span().Attribute("data-role", "samount").BindText(inEuros)),
			hx.Td().Body(bs.TextInput(l.notes)),
			hx.Td().Body(bs.SmallButton("outline-danger", "✕").Attribute("data-role", "remove-line").OnClick(func(hx.EventContext) {
				var kept []*lineState
				for _, other := range hx.UntrackGet(lines) {
					if other != l {
						kept = append(kept, other)
					}
				}
				lines.Set(kept)
			})))
	}

	sums := func(total *hx.ComputedT[int64], role string) hx.INode {
		span := hx.Span().Attribute("data-role", role)
		span.Effect(func() {
			span.Text(money(float64(total.Get())/100) + " €")
			span.ClassIf("text-success", balanced.Get()).ClassIf("text-danger", !balanced.Get())
		})
		return span
	}

	return hx.Div().Body(
		hx.Div().Class("table-responsive").Body(
			hx.Table().Class("table", "align-middle").Body(
				hx.THead().Body(hx.Tr().Body(
					hx.Th().Text(t.Get("Account")), hx.Th().Text(t.Get("Debit")), hx.Th().Text(t.Get("Credit")),
					hx.Th().Text(t.Get("Currency")), hx.Th().Class("text-end").Text(t.Get("InEuros")), hx.Th().Text(t.Get("Notes")), hx.Th())),
				hx.TBody().Body(hx.EachKeyed(lines, func(l *lineState) int { return l.key }, row)),
				hx.TFoot().Body(hx.Tr().Class("fw-bold").Body(
					hx.Td().Class("text-end").Text(t.Get("Total")+":"),
					hx.Td().Body(sums(debitSum, "sum-debit")),
					hx.Td().Body(sums(creditSum, "sum-credit")),
					hx.Td().Attribute("colspan", "4"))))),
		bs.Button("outline-primary", t.Get("NewLedgerLine")).Attribute("data-role", "add-line").OnClick(func(hx.EventContext) {
			l := newLine(Line{Account: int(ledger.AssetBank), Currency: string(currency.System)})
			lines.Set(append(append([]*lineState{}, hx.UntrackGet(lines)...), l))
		}))
}
