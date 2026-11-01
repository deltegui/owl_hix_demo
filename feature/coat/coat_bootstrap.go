// Package coat serves /coat, a page that shows every component of hxcoat. It is
// not part of the Bootstrap layout: it is a whole page of its own that loads
// Basecoat and Tailwind instead.
package coat

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/coat", handleShowIndex())
}
