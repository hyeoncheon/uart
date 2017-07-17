package actions

import (
	"github.com/gobuffalo/buffalo"
)

func t(c buffalo.Context, str string) string {
	s, err := T.Translate(c, str)
	if err == nil {
		if s == str {
			c.Logger().WithField("category", "i18n").
				Warnf("FIXME untranslated text found: %v", str)
		}
		return s
	}
	return str
}
