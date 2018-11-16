package actions

import (
	"html/template"
)

func imageForHelper(url, class string) template.HTML {
	return template.HTML(`<img class="` + class + `" src="` + url + `">`)
}

func logoForHelper(name string) template.HTML {
	fontName := map[string]string{
		"gplus":    "google",
		"github":   "github",
		"slack":    "slack",
		"facebook": "facebook-official",
	}
	return template.HTML(`<i class="fa fa-` + fontName[name] + `"></i>`)
}
