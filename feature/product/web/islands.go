package productweb

import (
	"github.com/deltegui/hx"
	"github.com/deltegui/owl/v3/island"

	"owldemo/web/ui/crud"
)

// The islands of the feature. The server places them in the page with
// island.Mount and the browser carries them with island.Run: the same
// definition builds them in both, from the same state, so the tree is the same.
var (
	ListIsland  = island.Define("products", func(s crud.State[Product]) hx.INode { return Screen(s) })
	StatsIsland = island.Define("products-stats", func(s StatsState) hx.INode { return Stats(s) })
)
