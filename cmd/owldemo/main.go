// Command owldemo is a CRUD app made with OWL (routes, middleware, validation,
// localization, CSRF) and Hix (the pages, rendered by the server, and the
// islands the browser hydrates). It is laid out like LittleCash: one folder per
// feature in feature/, each with its bootstrap, handler and service.
package main

import (
	"fmt"
	"time"

	"owldemo/feature/about"
	"owldemo/feature/contact"
	"owldemo/feature/entry"
	"owldemo/feature/forms"
	"owldemo/feature/language"
	"owldemo/feature/menu"
	"owldemo/feature/product"
	"owldemo/sys/bootstrap"
	"owldemo/sys/config"
	"owldemo/web/locale"
	"owldemo/web/static"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/owl/v3/csrf"
	"github.com/deltegui/owl/v3/cypher"
	"github.com/deltegui/owl/v3/middleware"
)

func main() {
	conf := config.CreateAndLoad()
	fmt.Println("owldemo", conf.Version)
	ctx := wire(conf)
	ctx.Mux.Listen(conf.ListenURL)
}

var modules = []bootstrap.BootstrapFunc{
	menu.Bootstrap,
	language.Bootstrap,
	about.Bootstrap,
	forms.Bootstrap,
	entry.Bootstrap,
	product.Bootstrap,
	contact.Bootstrap,
}

func wire(conf config.Configuration) bootstrap.Context {
	cy := cypher.NewWithPassword(conf.Secret)

	r := owl.New(cy)
	r.StaticEmbedded(static.Files)
	r.AddLocalization(locale.Files, locale.Shared, locale.Errors)

	// HixRenderer is the default. The templates of LittleCash would be
	// r.SetRenderer(owl.TemplateRenderer{}).

	r.Use(middleware.Logger)
	r.Use(middleware.Csrf(csrf.New(1*time.Hour, cy)))

	ctx := bootstrap.Context{Mux: r, Configuration: conf, Cypher: core.Cypher(cy)}
	for _, bootstrapFunc := range modules {
		bootstrapFunc(ctx)
	}
	return ctx
}
