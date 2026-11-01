// Command owldemo is a CRUD app made with OWL (routes, middleware, validation,
// localization, CSRF) and Hix (the pages, rendered by the server, and the
// islands the browser hydrates). It is laid out like LittleCash: one folder per
// feature in feature/, each with its bootstrap, handler and service.
package main

import (
	"fmt"
	"log"

	"owldemo/feature/about"
	"owldemo/feature/bulma"
	"owldemo/feature/bsdemo"
	"owldemo/feature/coat"
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
	"github.com/deltegui/owl/v3/cipher"
	"github.com/deltegui/owl/v3/csrf"
	"github.com/deltegui/owl/v3/middleware"
)

func main() {
	conf := config.CreateAndLoad()
	fmt.Println("owldemo", conf.Version)
	ctx, err := wire(conf)
	if err != nil {
		log.Fatalln("owldemo:", err)
	}
	ctx.Mux.Listen(conf.ListenURL)
}

var modules = []bootstrap.BootstrapFunc{
	menu.Bootstrap,
	language.Bootstrap,
	about.Bootstrap,
	bulma.Bootstrap,
	coat.Bootstrap,
	bsdemo.Bootstrap,
	forms.Bootstrap,
	entry.Bootstrap,
	product.Bootstrap,
	contact.Bootstrap,
}

func wire(conf config.Configuration) (bootstrap.Context, error) {
	cy, err := cipher.NewWithKeys(conf.Secret)
	if err != nil {
		return bootstrap.Context{}, err
	}

	r := owl.New(cy)
	r.StaticEmbedded(static.Files)
	r.AddLocalization(owl.Localization{Shared: locale.Shared, Errors: locale.Errors, Languages: locale.Languages})

	r.Use(middleware.Logger)
	// The demo is served over plain http, where a Secure cookie is not kept.
	r.Use(middleware.Csrf(csrf.New(cy, csrf.WithSecure(false))))

	ctx := bootstrap.Context{Mux: r, Configuration: conf, Cipher: cy}
	for _, bootstrapFunc := range modules {
		bootstrapFunc(ctx)
	}
	return ctx, nil
}
