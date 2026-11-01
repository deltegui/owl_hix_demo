package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/island"
	"github.com/deltegui/owl/v3/pagination"

	"owldemo/feature/product/web"
	"owldemo/sys/persistence"
	"owldemo/web/layout"
	"owldemo/web/ui/crud"
)

func filterOf(ctx owl.Ctx) productFilter {
	return productFilter{
		Name:     ctx.GetQueryParam("name"),
		Category: ctx.GetQueryParam("category"),
		Status:   ctx.GetQueryParam("status"),
	}
}

func pageOf(list pagination.List[productweb.Product]) crud.Page[productweb.Product] {
	p := list.Pagination
	return crud.Page[productweb.Product]{
		Items: list.Items,
		Page:  p.CurrentPage,
		Pages: max(1, (p.TotalElements+p.ElementsPerPage-1)/p.ElementsPerPage),
		Total: p.TotalElements,
	}
}

func encode(v any) string {
	text, _ := json.Marshal(v)
	return string(text)
}

// handleShowIndex renders the page: the chart and the list, each with its state.
func handleShowIndex(service productService) owl.Handler {
	return func(ctx owl.Ctx) error {
		list, err := service.Find(productFilter{}, pagination.Pagination{CurrentPage: 1})
		if err != nil {
			return ctx.StringInternalServerError("Error finding products")
		}
		stats, err := service.Stats()
		if err != nil {
			return ctx.StringInternalServerError("Error reading the stats")
		}
		page := pageOf(list)
		lang := ctx.GetCurrentLanguage()
		t := productweb.TextsFor(lang)
		return layout.Render(ctx, layout.Page{
			Title:  t.Get("Title"),
			Bundle: "full",
			Chart:  true,
			Islands: []layout.Island{
				layout.Of(island.Mount(productweb.StatsIsland, productweb.StatsState{Stats: stats, Lang: lang})),
				layout.Of(island.Mount(productweb.ListIsland, crud.State[productweb.Product]{Page: &page, Lang: lang})),
			},
		})
	}
}

func handleList(service productService) owl.Handler {
	return func(ctx owl.Ctx) error {
		current, _ := strconv.Atoi(ctx.GetQueryParam("page"))
		list, err := service.Find(filterOf(ctx), pagination.Pagination{CurrentPage: current})
		if err != nil {
			return ctx.StringInternalServerError("Error finding products")
		}
		return ctx.Json(pageOf(list))
	}
}

func handleStats(service productService) owl.Handler {
	return func(ctx owl.Ctx) error {
		stats, err := service.Stats()
		if err != nil {
			return ctx.StringInternalServerError("Error reading the stats")
		}
		return ctx.Json(stats)
	}
}

// handleSave creates the product, or updates it when the URL has an id.
func handleSave(service productService) owl.Handler {
	return func(ctx owl.Ctx) error {
		var item productweb.Product
		if err := ctx.ParseJson(&item); err != nil {
			return ctx.StringBadRequest("Invalid JSON")
		}
		id, ok := idOf(ctx)
		if !ok {
			return ctx.StringBadRequest("Invalid Id")
		}
		item.ID = id
		ctx.Validate(item)
		if !ctx.ModelState.Valid {
			return ctx.JsonInvalid(productweb.Texts)
		}
		saved, err := service.Save(item)
		if err != nil {
			return failure(ctx, err)
		}
		return ctx.Json(saved)
	}
}

func handleToggle(service productService) owl.Handler {
	return func(ctx owl.Ctx) error {
		id, ok := idOf(ctx)
		if !ok || id == 0 {
			return ctx.StringBadRequest("Invalid Id")
		}
		saved, err := service.Toggle(id)
		if err != nil {
			return failure(ctx, err)
		}
		return ctx.Json(saved)
	}
}

func handleDelete(service productService) owl.Handler {
	return func(ctx owl.Ctx) error {
		id, ok := idOf(ctx)
		if !ok || id == 0 {
			return ctx.StringBadRequest("Invalid Id")
		}
		if err := service.Delete(id); err != nil {
			return failure(ctx, err)
		}
		return ctx.NoContent()
	}
}

// idOf reads the :id of the URL. A URL with none is a new product, with id 0.
func idOf(ctx owl.Ctx) (int, bool) {
	raw := ctx.GetURLParam("id")
	if raw == "" {
		return 0, true
	}
	id, err := strconv.Atoi(raw)
	return id, err == nil
}

func failure(ctx owl.Ctx, err error) error {
	if errors.Is(err, persistence.ErrNotFound) {
		ctx.Status(http.StatusNotFound)
		return ctx.String("Not found")
	}
	return ctx.StringInternalServerError("Internal error")
}
