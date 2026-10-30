package forms

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	service := formsService{}
	ctx.Mux.Get("/forms/:name", handleShowIndex(service))
	ctx.Mux.Get("/api/forms/:name", handleDescribe(service))
}
