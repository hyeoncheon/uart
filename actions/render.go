package actions

// TODO REVIEW REQUIRED

import (
	"html/template"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/uuid"
)

var r *render.Engine
var assetsBox = packr.NewBox("../public")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Box containing all of the templates:
		//TemplatesBox: packr.NewBox("../templates"),
		TemplatesBox: packr.NewBox(uartHome + "/templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			"imageFor": imageForHelper,
			"logoFor":  logoForHelper,
			"iconize": func(s string) template.HTML {
				switch s {
				case "admin":
					return template.HTML(`<i class="fa fa-empire"></i>`)
				default:
					return template.HTML(`<i class="fa fa-` + s + `"></i>`)
				}
			},
			"trunc": func(t interface{}, args ...int) string {
				length := 20
				var s string
				switch t.(type) {
				case string:
					s = t.(string)
				case uuid.UUID:
					s = t.(uuid.UUID).String()
					length = 14
				}

				if len(args) > 0 {
					length = args[0]
				}
				if length > len(s)-4 {
					return s
				}
				return s[0:length] + "..."
			},
		},
	})
}
