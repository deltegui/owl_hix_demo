package menu

import (
	"owldemo/sys/bootstrap"
	"owldemo/web/layout"
)

func Bootstrap(ctx bootstrap.Context) {
	layout.Menu = getMenu(menuService{})
}
