// Package productweb is what the product feature shows, in the browser and on
// the server: the data, the screen and the chart. It builds on both sides (the
// server renders the islands and the browser hydrates them), so it imports only
// Hix and the UI packages, never OWL.
//
// The valtruc tags are only read by the server.
package productweb

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name" valtruc:"min=3, max=60, required"`
	Category string  `json:"category" valtruc:"required"`
	Price    float64 `json:"price" valtruc:"min=0.01, max=100000"`
	Active   bool    `json:"active"`
}

var Categories = []string{"Books", "Games", "Music", "Tools"}

// CategoryStat is what the chart shows.
type CategoryStat struct {
	Category string  `json:"category"`
	Count    int     `json:"count"`
	Value    float64 `json:"value"`
}

// StatsState is the state of the chart island.
type StatsState struct {
	Stats []CategoryStat `json:"stats"`
	Lang  string         `json:"l"`
}
