package entry

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	service := newEntryService()

	pageMux := ctx.Mux.CreateSubMux("/entries")
	pageMux.Get("/", handleShowIndex(service))

	// httprouter does not mix /entries/new with /entries/:id, so the blank entry
	// has a route of its own.
	ctx.Mux.Get("/api/entry-defaults", handleShowCreate())

	apiMux := ctx.Mux.CreateSubMux("/api/entries")
	apiMux.Get("/", handleList(service))
	apiMux.Get("/:id", handleShowOne(service))
	apiMux.Post("/", handleSave(service))
	apiMux.Put("/:id", handleSave(service))
	apiMux.Delete("/:id", handleDelete(service))
	apiMux.Post("/line", handleLine(service))
	apiMux.Post("/convert", handleConvert())
}
