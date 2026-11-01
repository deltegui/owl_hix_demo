// Package bulma serves /bulma, a page that shows every component of hxbulma.
// It is not part of the Bootstrap layout: it is a whole page of its own that
// loads Bulma instead.
package bulma

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/bulma", handleShowIndex())
}
