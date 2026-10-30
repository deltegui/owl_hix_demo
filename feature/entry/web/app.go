package entryweb

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/deltegui/hx"

	"owldemo/sys/currency"
	"owldemo/web/ui/crud"
)

// layer is what the second layer of the page shows, over the list: nothing, the
// editor of an entry or its viewer. ID 0 is a new entry. (LittleCash has two
// layers too: main-layer is the list and alternate-layer the editor.)
type layer struct {
	Kind string // "", "edit" or "view"
	ID   int
}

// app is what the screens of the island share.
type app struct {
	t    crud.Texts
	look Lookups

	notice, failure *hx.SignalT[string]
	open            *hx.SignalT[layer]
	loaded          *hx.SignalT[*Entry] // the entry the second layer shows, once the server answered
	loading         *hx.ActionT[int, Entry]
	reload          func() // loads the list again
}

func newApp(t crud.Texts, look Lookups) *app {
	a := &app{t: t, look: look,
		notice: hx.Signal(""), failure: hx.Signal(""),
		open: hx.Signal(layer{}), loaded: hx.Signal[*Entry](nil),
		reload: func() {},
	}
	a.loading = hx.Action(func(ctx context.Context, id int) (Entry, error) { return fetchEntry(ctx, id) })
	a.loading.OnSuccess(func(e Entry, _ int) { a.loaded.Set(&e) })
	a.loading.OnError(func(err error, _ int) {
		a.close()
		a.failure.Set(err.Error())
	})
	return a
}

// show opens the second layer for an entry and asks the server for it.
func (a *app) show(kind string, id int) {
	a.notice.Set("")
	a.failure.Set("")
	a.loaded.Set(nil)
	a.open.Set(layer{Kind: kind, ID: id})
	a.loading.Run(id)
}

func (a *app) close() {
	a.open.Set(layer{})
	a.loaded.Set(nil)
}

// getter turns a function into something that can be read, so it can be passed
// where a signal is expected. The reads inside it are tracked by whoever calls Get.
type getter[T any] func() T

func (g getter[T]) Get() T { return g() }

func money(v float64) string { return fmt.Sprintf("%.2f", v) }

func parseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func itoa(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// cents is an amount in a currency as the system currency, in cents, like the
// server does it (see currency.ToSystem).
func cents(amount, rate float64) int64 {
	return int64(math.Round(float64(currency.ToCents(amount)) * rate))
}

func (a *app) account(code int) (AccountInfo, bool) {
	for _, acc := range a.look.Accounts {
		if acc.Code == code {
			return acc, true
		}
	}
	return AccountInfo{}, false
}

func (a *app) operation(id int) (OperationInfo, bool) {
	for _, op := range a.look.Operations {
		if op.ID == id {
			return op, true
		}
	}
	return OperationInfo{}, false
}

func (a *app) categoryName(id int) string {
	for _, c := range a.look.Categories {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

// accountsOf are the accounts of a group, or all of them if group is 0.
func (a *app) accountsOf(group int) []AccountInfo {
	var out []AccountInfo
	for _, acc := range a.look.Accounts {
		if group == 0 || acc.Group == group {
			out = append(out, acc)
		}
	}
	return out
}

// fullAccountName is "Assets · Bank".
func (a *app) fullAccountName(code int) string {
	acc, ok := a.account(code)
	if !ok {
		return strconv.Itoa(code)
	}
	return a.t.Get("Group"+strconv.Itoa(acc.Group)) + " · " + a.t.Get(acc.Name)
}
