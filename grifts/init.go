package grifts

import (
	"github.com/gobuffalo/buffalo"

	"github.com/hyeoncheon/uart/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
