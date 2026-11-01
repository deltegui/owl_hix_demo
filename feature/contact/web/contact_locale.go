package contactweb

import (
	"github.com/deltegui/owl/v3/localizer"

	"owldemo/web/locale"
	"owldemo/web/ui/crud"
)

// Texts is the translations of the page. The server and the browser both have
// them, so a screen can be built in either.
var Texts = localizer.Dict{
	localizer.LangSpanish: {
		"Title":          "Contactos",
		"Subtitle":       "La misma pantalla, configurada para otra entidad.",
		"FormTitle":      "Contacto",
		"NewButton":      "Nuevo contacto",
		"DeleteTitle":    "Eliminar contacto",
		"DeleteQuestion": "¿Eliminar %s? No se puede deshacer.",
		"SavedMessage":   "Contacto guardado.",
		"DeletedMessage": "Contacto eliminado.",
		"#":              "#",
		"Name":           "Nombre",
		"Email":          "Correo",
		"Phone":          "Teléfono",
		"Status":         "Estado",
		"Active":         "Activo",
		"Query":          "Nombre o correo",
		"emailAt":        "el correo",
	},
	localizer.LangEnglish: {
		"Title":          "Contacts",
		"Subtitle":       "The same screen, configured for another entity.",
		"FormTitle":      "Contact",
		"NewButton":      "New contact",
		"DeleteTitle":    "Delete contact",
		"DeleteQuestion": "Delete %s? This cannot be undone.",
		"SavedMessage":   "Contact saved.",
		"DeletedMessage": "Contact deleted.",
		"#":              "#",
		"Name":           "Name",
		"Email":          "Email",
		"Phone":          "Phone",
		"Status":         "Status",
		"Active":         "Active",
		"Query":          "Name or email",
		"emailAt":        "the email",
	},
}

// TextsFor is what the page says in the language: the shared translations, and
// Texts over them.
func TextsFor(lang string) crud.Texts {
	return crud.TextsOf(lang, locale.Languages[0], locale.Shared, Texts)
}
