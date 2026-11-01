package menu

import (
	"testing"

	"github.com/deltegui/owl/v3/localizer"

	"owldemo/web/locale"
)

func TestTexts(t *testing.T) {
	for _, problem := range localizer.Check(viewMenu, locale.Languages...) {
		t.Error(problem)
	}
}
