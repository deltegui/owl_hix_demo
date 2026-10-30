package language

import (
	"net/http"
	"strings"

	"github.com/deltegui/owl/v3"
)

// handleChange sets the language cookie and goes back to the page the user was
// in (the form sends it in Next). The CSRF middleware has already checked the token.
func handleChange() owl.Handler {
	return func(ctx owl.Ctx) error {
		if err := ctx.ChangeLanguage(ctx.GetURLParam("lang")); err != nil {
			return ctx.StringInternalServerError("Cannot change the language")
		}
		return ctx.RedirectCode(localPath(ctx.Req.FormValue("Next")), http.StatusSeeOther)
	}
}

// localPath keeps only paths of this site: "//evil.com" and "https://evil.com"
// would make the form an open redirect.
func localPath(next string) string {
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.HasPrefix(next, "/\\") {
		return "/"
	}
	return next
}
