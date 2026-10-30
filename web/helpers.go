// Package web has what the features share to answer: JSON for the islands.
package web

import (
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/valtruc"

	"owldemo/web/ui/crud"
)

// Invalid answers 422 with what ctx.Validate found, by field and translated with
// the file of the view (the field names come from there). The fields are named
// like the JSON of the model, which is what the island knows.
//
// A missing value is the only thing worth saying about a field: the other rules
// would only repeat it.
func Invalid(ctx owl.Ctx, file string) error {
	vm := ctx.ViewModel(file, nil)
	errs := map[string][]string{}
	for field, list := range ctx.ModelState.Errors {
		name := jsonPath(field)
		for _, err := range list {
			if err.GetIdentifier() == valtruc.RequiredIdentifier {
				errs[name] = []string{vm.FormatError(err)}
				break
			}
			errs[name] = append(errs[name], vm.FormatError(err))
		}
	}
	ctx.Status(http.StatusUnprocessableEntity)
	return ctx.Json(crud.Errors{Errors: errs})
}

// jsonPath names a field like the JSON does: "Lines[0].Account" is "lines[0].account".
func jsonPath(path string) string {
	parts := strings.Split(path, ".")
	for i, part := range parts {
		parts[i] = lowerFirst(part)
	}
	return strings.Join(parts, ".")
}

func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

// Failure answers 400 with the message of a rule of the domain, translated with
// the errors file (by the code of the error), for the island to show.
func Failure(ctx owl.Ctx, err core.DomainError) error {
	ctx.Status(http.StatusBadRequest)
	return ctx.Json(crud.Errors{Message: ctx.LocalizeError(err)})
}
