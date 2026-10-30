package about

import (
	"github.com/deltegui/owl/v3"

	"owldemo/sys/bootstrap"
)

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Get("/about", handleShowIndex())
	ctx.Mux.Get("/", func(c owl.Ctx) error { return c.Redirect("/products") })
}
