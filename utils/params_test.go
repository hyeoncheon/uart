package utils_test

import (
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/markbates/willie"
	"github.com/stretchr/testify/require"

	"github.com/hyeoncheon/uart/utils"
)

func Test_GetParam(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		xstr := utils.GetParam(c, "string")
		xint := utils.GetParam(c, "integer")
		return c.Render(200, render.String("%s-%d", xstr, xint))
	})
	w := willie.New(a)

	res := w.Request("/?string=value&integer=1").Get()
	r.Equal("value-1", res.Body.String())
}

func Test_GetIntParam(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		val := utils.GetIntParam(c, "val", 1, 10)
		max := utils.GetIntParam(c, "max", 1, 10)
		min := utils.GetIntParam(c, "min", 1, 10)
		return c.Render(200, render.String("%d-%d-%d", val, max, min))
	})
	w := willie.New(a)

	res := w.Request("/?val=5&max=11&min=0").Get()
	r.Equal("5-10-1", res.Body.String())

	res = w.Request("/?val=5&max=8&min=2").Get()
	r.Equal("5-8-2", res.Body.String())
}

func Test_GetStringParam(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		xstr := utils.GetStringParam(c, "param", "default")
		return c.Render(200, render.String("%s", xstr))
	})
	w := willie.New(a)

	res := w.Request("/?param=value").Get()
	r.Equal("value", res.Body.String())

	res = w.Request("/?str=value").Get()
	r.Equal("default", res.Body.String())
}