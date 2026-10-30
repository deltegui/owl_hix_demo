package crud

import "fmt"

// Texts are the translations a screen shows: the localizer of the view, which the
// server sends in the state of the island, so the browser writes the same words
// the server did. A key that is not there is shown as it is.
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
