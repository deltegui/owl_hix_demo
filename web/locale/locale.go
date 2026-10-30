// Package locale has the translations, one json file per view (like
// views/<feature>/<view>.html in LittleCash) and the shared ones.
package locale

import "embed"

//go:embed *
var Files embed.FS

const (
	Shared string = "common/shared"
	Errors string = "common/errors"
)
