package locale

import (
	"testing"

	"github.com/deltegui/owl/v3/localizer"
)

func TestTexts(t *testing.T) {
	for name, dict := range map[string]localizer.Dict{"Shared": Shared, "Errors": Errors} {
		for _, problem := range localizer.Check(dict, Languages...) {
			t.Errorf("%s: %s", name, problem)
		}
	}
}
