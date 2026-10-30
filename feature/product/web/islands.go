package productweb

import (
	"encoding/json"

	"github.com/deltegui/hx"

	"owldemo/web/ui/crud"
)

// The islands of the feature, as the registries want them: they build from the
// state the server sent (JSON). The server calls them to render the island and
// the browser calls them to hydrate it, with the same text, so the tree is the same.

func decode[T any](state string) T {
	var v T
	if state != "" {
		json.Unmarshal([]byte(state), &v)
	}
	return v
}

func ListIsland(state string) hx.INode  { return Screen(decode[crud.State[Product]](state)) }
func StatsIsland(state string) hx.INode { return Stats(decode[StatsState](state)) }
