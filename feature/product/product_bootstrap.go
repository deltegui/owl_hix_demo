package product

import "owldemo/sys/bootstrap"

func Bootstrap(ctx bootstrap.Context) {
	service := newProductService()

	pageMux := ctx.Mux.CreateSubMux("/products")
	pageMux.Get("/", handleShowIndex(service))

	apiMux := ctx.Mux.CreateSubMux("/api/products")
	apiMux.Get("/", handleList(service))
	apiMux.Get("/stats", handleStats(service))
	apiMux.Post("/", handleSave(service))
	apiMux.Put("/:id", handleSave(service))
	apiMux.Delete("/:id", handleDelete(service))
	apiMux.Post("/:id/toggle", handleToggle(service))
}
