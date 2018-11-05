package utils_test

import (
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/require"

	"github.com/hyeoncheon/uart/utils"
)

func Test_A_OOPS(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		return utils.OOPS(c, 302, "/302", "test", "mesg %v", "value")
	})
	w := httptest.New(a)

	res := w.HTML("/").Get()
	r.Equal(302, res.Code)
	r.Equal("/302", res.HeaderMap.Get("Location"))
}

func Test_B_DOOPS(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		return utils.DOOPS(c, "mesg %v", "value")
	})
	w := httptest.New(a)

	res := w.HTML("/").Get()
	r.Equal(302, res.Code)
	r.Equal("/", res.HeaderMap.Get("Location"))
}

func Test_C_SOOPS(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		return utils.SOOPS(c, "mesg %v", "value")
	})
	w := httptest.New(a)

	res := w.HTML("/").Get()
	r.Equal(302, res.Code)
	r.Equal("/", res.HeaderMap.Get("Location"))
}

func Test_D_InvalidAccess(t *testing.T) {
	r := require.New(t)
	a := buffalo.New(buffalo.Options{})

	a.GET("/", func(c buffalo.Context) error {
		return utils.InvalidAccess(c, "/hell", "mesg %v", "value")
	})
	w := httptest.New(a)

	res := w.HTML("/").Get()
	r.Equal(302, res.Code)
	r.Equal("/hell", res.HeaderMap.Get("Location"))
}
