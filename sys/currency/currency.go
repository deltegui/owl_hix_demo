// Package currency has the currencies of the demo and a daily exchange rate to
// the system currency. The rates are made up (the same on every run), but they
// depend on the date, so an entry keeps the rate of the day it uses.
//
// Amounts are int64 in cents. It imports nothing, so it builds for the browser.
package currency

import (
	"math"
	"time"
)

type Currency string

const (
	EUR Currency = "EUR"
	USD Currency = "USD"
	GBP Currency = "GBP"
)

// System is the currency every amount is converted to (the "samount").
const System = EUR

func All() []Currency { return []Currency{EUR, USD, GBP} }

func (c Currency) Sign() string {
	switch c {
	case EUR:
		return "€"
	case USD:
		return "$"
	case GBP:
		return "£"
	}
	return string(c)
}

func (c Currency) Valid() bool {
	for _, known := range All() {
		if c == known {
			return true
		}
	}
	return false
}

// ToCents turns 12.34 into 1234.
func ToCents(amount float64) int64 { return int64(math.Round(amount * 100)) }

// FromCents turns 1234 into 12.34.
func FromCents(cents int64) float64 { return float64(cents) / 100 }

// LastExchangeDate is the day of the newest rates: yesterday, at midnight UTC.
func LastExchangeDate() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
}

// Rate is how many euros a unit of the currency is worth on that day. It moves a
// little with the day of the year, like a market would.
func Rate(c Currency, date time.Time) float64 {
	var base, swing float64
	switch c {
	case USD:
		base, swing = 0.92, 0.0004
	case GBP:
		base, swing = 1.17, 0.0006
	default:
		return 1
	}
	wave := float64(date.YearDay()%20) - 10
	return math.Round((base+swing*wave)*10000) / 10000
}

// ToSystem converts cents of a currency to cents of the system currency with the
// rate of the day, and returns the rate used.
func ToSystem(cents int64, c Currency, date time.Time) (system int64, rate float64) {
	rate = Rate(c, date)
	return int64(math.Round(float64(cents) * rate)), rate
}
