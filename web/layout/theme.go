package layout

import (
	"github.com/deltegui/hx"

	"owldemo/web/ui/theme"
)

func themeBuild(label string) hx.INode { return theme.Build(label) }
