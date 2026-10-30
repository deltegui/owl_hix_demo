package contact

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	service := newContactService()

	pageMux := ctx.Mux.CreateSubMux("/contacts")
	pageMux.Get("/", handleShowIndex(service))

	apiMux := ctx.Mux.CreateSubMux("/api/contacts")
	apiMux.Get("/", handleList(service))
	apiMux.Post("/", handleSave(service))
	apiMux.Put("/:id", handleSave(service))
	apiMux.Delete("/:id", handleDelete(service))
	apiMux.Post("/:id/toggle", handleToggle(service))
}
