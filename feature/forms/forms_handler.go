package forms

import (
	"encoding/json"
	"net/http"

	"github.com/deltegui/owl/v3"

	formsweb "owldemo/feature/forms/web"
	"owldemo/web/layout"
	"owldemo/web/ui/crud"
)

const viewFormsIndex string = "forms/forms_index"

// handleShowIndex renders the page of a form: the form is drawn by the server
// and its description goes in the state, so the browser has nothing to load.
func handleShowIndex(service formsService) owl.Handler {
	return func(ctx owl.Ctx) error {
		form, ok := service.find(ctx.GetURLParam("name"))
		if !ok {
			ctx.Status(http.StatusNotFound)
			return ctx.String("Not found")
		}
		t := crud.Texts(ctx.GetLocalizer(viewFormsIndex))
		state, _ := json.Marshal(formsweb.State{Form: form, Texts: t})
		return layout.Render(ctx, layout.Page{
			Title:   t.Get(form.Title),
			Bundle:  "forms",
			Islands: []layout.Island{{Name: "form", State: string(state), Build: formsweb.Island}},
		})
	}
}

// handleDescribe serves the description of a form, for clients that do not get it with the page.
func handleDescribe(service formsService) owl.Handler {
	return func(ctx owl.Ctx) error {
		form, ok := service.find(ctx.GetURLParam("name"))
		if !ok {
			ctx.Status(http.StatusNotFound)
			return ctx.String("Not found")
		}
		return ctx.Json(form)
	}
}
