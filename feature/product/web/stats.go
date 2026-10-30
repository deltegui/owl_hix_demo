package productweb

import (
	"context"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/hxjson"

	"owldemo/web/ui/bs"
	"owldemo/web/ui/chart"
)

// Stats is the island with the Chart.js chart. It starts from the stats the
// server rendered, reads its own endpoint when it has none, and reads it again
// whenever the products change in another island (Version).
func Stats(state StatsState) hx.INode {
	load := func(ctx context.Context, _ int) ([]CategoryStat, error) {
		return hxjson.Get[[]CategoryStat](ctx, "/api/products/stats")
	}
	var stats *hx.ResourceT[[]CategoryStat]
	if state.Stats != nil {
		stats = hx.ResourceForWith(state.Stats, Version, load)
	} else {
		stats = hx.ResourceFor(Version, load)
	}
	data := getData(func() chart.Data {
		var d chart.Data
		for _, s := range stats.Value().Get() {
			d.Labels = append(d.Labels, state.Texts.Or("Opt."+s.Category, s.Category))
			d.Values = append(d.Values, s.Value)
		}
		return d
	})
	return bs.Card(state.Texts.Get("StatsTitle"), chart.New("bar", state.Texts.Get("StatsLabel"), data))
}

// getData is a function that can be read like a signal.
type getData func() chart.Data

func (g getData) Get() chart.Data { return g() }
