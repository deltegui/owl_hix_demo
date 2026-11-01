// Package bsdemo serves /bs, a page that shows every component of hxbs. Like
// /coat it is a whole page of its own: it loads Bootstrap from the CDN.
package bsdemo

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/bs", handleShowIndex())
}
