package productweb

import (
	"github.com/deltegui/owl/v3/localizer"

	"owldemo/web/locale"
	"owldemo/web/ui/crud"
)

// Texts is the translations of the page. The server and the browser both have
// them, so a screen can be built in either.
var Texts = localizer.Dict{
	localizer.LangSpanish: {
		"Title":          "Productos",
		"Subtitle":       "Filtra, crea, edita, habilita y elimina productos.",
		"FormTitle":      "Producto",
		"NewButton":      "Nuevo producto",
		"DeleteTitle":    "Eliminar producto",
		"DeleteQuestion": "¿Eliminar %s? No se puede deshacer.",
		"SavedMessage":   "Producto guardado.",
		"DeletedMessage": "Producto eliminado.",
		"#":              "#",
		"Name":           "Nombre",
		"Category":       "Categoría",
		"Price":          "Precio",
		"Status":         "Estado",
		"Active":         "Activo",
		"StatsTitle":     "Productos activos: precio total por categoría",
		"StatsLabel":     "Precio total",
		"Opt.Books":      "Libros",
		"Opt.Games":      "Juegos",
		"Opt.Music":      "Música",
		"Opt.Tools":      "Herramientas",
	},
	localizer.LangEnglish: {
		"Title":          "Products",
		"Subtitle":       "Filter, create, edit, enable and delete products.",
		"FormTitle":      "Product",
		"NewButton":      "New product",
		"DeleteTitle":    "Delete product",
		"DeleteQuestion": "Delete %s? This cannot be undone.",
		"SavedMessage":   "Product saved.",
		"DeletedMessage": "Product deleted.",
		"#":              "#",
		"Name":           "Name",
		"Category":       "Category",
		"Price":          "Price",
		"Status":         "Status",
		"Active":         "Active",
		"StatsTitle":     "Active products: total price by category",
		"StatsLabel":     "Total price",
		"Opt.Books":      "Books",
		"Opt.Games":      "Games",
		"Opt.Music":      "Music",
		"Opt.Tools":      "Tools",
	},
}

// TextsFor is what the page says in the language: the shared translations, and
// Texts over them.
func TextsFor(lang string) crud.Texts {
	return crud.TextsOf(lang, locale.Languages[0], locale.Shared, Texts)
}
