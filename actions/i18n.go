package actions

// TODO REVIEW REQUIRED

import (
	"github.com/gobuffalo/buffalo"
)

func t(c buffalo.Context, str string) string {
	s := T.Translate(c, str)
	if s == str {
		c.Logger().WithField("category", "i18n").Warnf("UNTRANSLATED: %v", str)
	}
	return s
}
