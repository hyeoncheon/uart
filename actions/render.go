package actions

// TODO REVIEW REQUIRED

import (
	"time"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Box containing all of the templates:
		//TemplatesBox: packr.NewBox("../templates"),
		TemplatesBox: packr.NewBox(uartHome + "/templates"),

		// Add template helpers here:
		Helpers: render.Helpers{
			"shorten":  shortenHelper,
			"imageFor": imageForHelper,
			"logoFor":  logoForHelper,
			"timeYYMDHMS": func(t time.Time) string {
				return t.Local().Format("2006-01-02 15:04:05")
			},
			"timeYYMDHM": func(t time.Time) string {
				return t.Local().Format("2006-01-02 15:04")
			},
			"timeYMDHM": func(t time.Time) string {
				return t.Local().Format("06-01-02 15:04")
			},
			"timeMDHM": func(t time.Time) string {
				return t.Local().Format("01-02 15:04")
			},
			"timeYYMD": func(t time.Time) string {
				return t.Local().Format("2006-01-02")
			},
		},
	})
}
