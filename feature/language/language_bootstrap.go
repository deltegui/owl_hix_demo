package language

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	ctx.Mux.Post("/language/:lang", handleChange())
}
