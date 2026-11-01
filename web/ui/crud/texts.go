package crud

import (
	"fmt"
	"maps"

	"github.com/deltegui/owl/v3/localizer"
)

// Texts are the translations a screen shows: the words of the page in a
// language. The server and the browser each build them with TextsOf from the same
// dictionaries and the same language (the state of the island carries the
// language), so the browser writes the same words the server did. A key that is
// not there is shown as it is.
//
// The screen uses these keys (the shared file has the ones that start with
// "Shared."): Title, Subtitle, FormTitle, NewButton, DeleteTitle, DeleteQuestion
// (%s is the name), SavedMessage, DeletedMessage, the label of every column,
// filter and field, "Opt.<value>" for the text of an option, and Shared.Filter,
// Shared.Search, Shared.Clear, Shared.All, Shared.Choose, Shared.Edit,
// Shared.Delete, Shared.Enable, Shared.Disable, Shared.Cancel, Shared.Save,
// Shared.Saving, Shared.Active, Shared.Disabled, Shared.Loading,
// Shared.NothingFound, Shared.Found (%d) and Shared.LoadError (%s).
type Texts map[string]string

// TextsOf is the words of the sources in the language, one over the other: the
// last source wins when a key is in several. A screen says its own words over the
// shared ones: TextsOf(lang, fallback, shared, own). A source with nothing in the
// language is read in fallback, the default language of the application.
func TextsOf(lang, fallback string, sources ...localizer.Source) Texts {
	t := Texts{}
	for _, src := range sources {
		maps.Copy(t, localizer.Pick(src, lang, fallback))
	}
	return t
}

// Get is the text of the key, formatted with args if there are any.
func (t Texts) Get(key string, args ...any) string {
	text, ok := t[key]
	if !ok {
		return key
	}
	if len(args) == 0 {
		return text
	}
	return fmt.Sprintf(text, args...)
}

// Or is the text of the key, or fallback when there is none.
func (t Texts) Or(key, fallback string) string {
	if text, ok := t[key]; ok {
		return text
	}
	return fallback
}
