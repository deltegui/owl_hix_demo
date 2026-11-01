package formsweb

import (
	"github.com/deltegui/owl/v3/localizer"

	"owldemo/web/locale"
	"owldemo/web/ui/crud"
)

// Texts is the translations of the page. The server and the browser both have
// them, so a screen can be built in either.
var Texts = localizer.Dict{
	localizer.LangSpanish: {
		"Form.product": "Nuevo producto",
		"Form.contact": "Nuevo contacto",
		"Saved":        "Guardado con el id %v.",
		"Name":         "Nombre",
		"Category":     "Categoría",
		"Price":        "Precio",
		"Active":       "Activo",
		"Email":        "Correo",
		"Phone":        "Teléfono",
		"Opt.Books":    "Libros",
		"Opt.Games":    "Juegos",
		"Opt.Music":    "Música",
		"Opt.Tools":    "Herramientas",
	},
	localizer.LangEnglish: {
		"Form.product": "New product",
		"Form.contact": "New contact",
		"Saved":        "Saved with id %v.",
		"Name":         "Name",
		"Category":     "Category",
		"Price":        "Price",
		"Active":       "Active",
		"Email":        "Email",
		"Phone":        "Phone",
		"Opt.Books":    "Books",
		"Opt.Games":    "Games",
		"Opt.Music":    "Music",
		"Opt.Tools":    "Tools",
	},
}

// TextsFor is what the page says in the language: the shared translations, and
// Texts over them.
func TextsFor(lang string) crud.Texts {
	return crud.TextsOf(lang, locale.Languages[0], locale.Shared, Texts)
}
