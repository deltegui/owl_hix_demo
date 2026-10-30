package contact

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/pagination"

	"owldemo/feature/contact/web"
	"owldemo/sys/persistence"
	"owldemo/web"
	"owldemo/web/layout"
	"owldemo/web/ui/crud"
)

const viewContactIndex string = "contact/contact_index"

func filterOf(ctx owl.Ctx) contactFilter {
	return contactFilter{
		Query:  ctx.GetQueryParam("q"),
		Status: ctx.GetQueryParam("status"),
	}
}

func pageOf(list pagination.List[contactweb.Contact]) crud.Page[contactweb.Contact] {
	p := list.Pagination
	return crud.Page[contactweb.Contact]{
		Items: list.Items,
		Page:  p.CurrentPage,
		Pages: max(1, (p.TotalElements+p.ElementsPerPage-1)/p.ElementsPerPage),
		Total: p.TotalElements,
	}
}

func texts(ctx owl.Ctx) crud.Texts { return crud.Texts(ctx.GetLocalizer(viewContactIndex)) }

func encode(v any) string {
	text, _ := json.Marshal(v)
	return string(text)
}

// handleShowIndex renders the page: the list, with its state.
func handleShowIndex(service contactService) owl.Handler {
	return func(ctx owl.Ctx) error {
		list, err := service.Find(contactFilter{}, pagination.Pagination{CurrentPage: 1})
		if err != nil {
			return ctx.StringInternalServerError("Error finding contacts")
		}
		page := pageOf(list)
		t := texts(ctx)
		return layout.Render(ctx, layout.Page{
			Title:  t.Get("Title"),
			Bundle: "full",
			Islands: []layout.Island{
				{Name: "contacts", State: encode(crud.State[contactweb.Contact]{Page: &page, Texts: t}), Build: contactweb.ListIsland},
			},
		})
	}
}

func handleList(service contactService) owl.Handler {
	return func(ctx owl.Ctx) error {
		current, _ := strconv.Atoi(ctx.GetQueryParam("page"))
		list, err := service.Find(filterOf(ctx), pagination.Pagination{CurrentPage: current})
		if err != nil {
			return ctx.StringInternalServerError("Error finding contacts")
		}
		return ctx.Json(pageOf(list))
	}
}

// handleSave creates the contact, or updates it when the URL has an id.
func handleSave(service contactService) owl.Handler {
	return func(ctx owl.Ctx) error {
		var item contactweb.Contact
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
			return web.Invalid(ctx, viewContactIndex)
		}
		saved, err := service.Save(item)
		if err != nil {
			return failure(ctx, err)
		}
		return ctx.Json(saved)
	}
}

func handleToggle(service contactService) owl.Handler {
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

func handleDelete(service contactService) owl.Handler {
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

// idOf reads the :id of the URL. A URL with none is a new contact, with id 0.
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
