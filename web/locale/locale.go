// Package locale has the translations that belong to the whole application: the
// shared ones and the errors, and the languages it speaks. The ones of a page or
// a component are a localizer.Dict next to its code.
//
// The words of the pagination bar are pagination.Texts, in Spanish and English.
package locale

import "github.com/deltegui/owl/v3/localizer"

// Languages are the ones the application speaks. The first one is the default.
// The server gives them to AddLocalization and the browser uses them to pick the
// words of a page, so both say the same.
var Languages = []string{localizer.LangSpanish, localizer.LangEnglish}
