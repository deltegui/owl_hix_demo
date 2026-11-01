package about

import "github.com/deltegui/owl/v3/localizer"

// viewAboutIndex is the translations of the page.
var viewAboutIndex = localizer.Dict{
	localizer.LangSpanish: {
		"Title":    "Acerca de",
		"Subtitle": "Una demo de OWL con Hix",
		"Body1":    "Esta página es solo HTML: la dibuja el servidor con Hix y no hay nada que hidratar, así que carga el WASM pequeño (solo el selector de tema).",
		"Body2":    "El resto de pantallas son islas: el servidor las dibuja con su estado y el navegador las hidrata.",
		"Features": "Una carpeta por funcionalidad, como en LittleCash: bootstrap, handler y service, con sus vistas y sus textos.",
	},
	localizer.LangEnglish: {
		"Title":    "About",
		"Subtitle": "An OWL demo with Hix",
		"Body1":    "This page is only HTML: the server draws it with Hix and there is nothing to hydrate, so it loads the small WASM (just the theme selector).",
		"Body2":    "The other screens are islands: the server draws them with their state and the browser hydrates them.",
		"Features": "One folder per feature, like LittleCash: bootstrap, handler and service, with its views and its texts.",
	},
}
