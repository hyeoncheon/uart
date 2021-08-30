package actions

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Helpers_ImageForHelper(t *testing.T) {
	r := require.New(t)
	r.Equal(template.HTML(`<img class="class" src="url">`), imageForHelper("url", "class"))
}

func Test_Helpers_LogoForHelper(t *testing.T) {
	r := require.New(t)
	r.Equal(template.HTML(`<i class="fab fa-google"></i>`), logoForHelper("gplus"))
	r.Equal(template.HTML(`<i class="fab fa-facebook"></i>`), logoForHelper("facebook"))
	r.Equal(template.HTML(`<i class="fab fa-github"></i>`), logoForHelper("github"))
}
