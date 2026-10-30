// Package bootstrap is what every feature gets to wire itself: the multiplexer
// and what it needs from the system. Each feature has a Bootstrap(Context) that
// registers its routes; cmd/owldemo lists them.
package bootstrap

import (
	"owldemo/sys/config"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/core"
)

type BootstrapFunc func(Context)

type Context struct {
	Mux           *owl.Mux
	Configuration config.Configuration
	Cypher        core.Cypher
}
