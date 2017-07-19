package actions

import (
	"fmt"
	"html/template"
	"net/url"

	"github.com/hyeoncheon/uart/models"
)

func imageForHelper(url, class string) template.HTML {
	return template.HTML(`<img class="` + class + `" src="` + url + `">`)
}

func paginateHelper(sp models.SearchParams) template.HTML {
	var str string
	pagerLen := 11
	center := pagerLen/2 + 1
	arm := pagerLen/2 - 2

	loopStart := 1
	loogEnd := sp.TotalPages

	query := ""
	if sp.Sort != "" {
		query += "&sort=" + sp.Sort
	}
	if sp.FilterKey != "" {
		value := url.QueryEscape(fmt.Sprintf("%v", sp.FilterValue))
		query += fmt.Sprintf("&filter=%v&value=%v", sp.FilterKey, value)
	}

	if sp.TotalPages > pagerLen {
		loogEnd = pagerLen - 2
		if sp.Page > center {
			loopStart = sp.Page - arm
			loogEnd = sp.Page + arm
			str += fmt.Sprintf(`<li><a href="?page=1&pp=%v%v">1</a></li>`,
				sp.PerPage, query)
			str += `<li><a>...</a></li>`
		}
		if sp.Page > (sp.TotalPages - arm - 3) {
			loogEnd = sp.TotalPages
			loopStart = sp.TotalPages - pagerLen + 3
		}
	}
	for i := loopStart; i <= loogEnd; i++ {
		attr := ""
		if i == sp.Page {
			attr = ` class="active"`
		}
		str += fmt.Sprintf(`<li%v><a href="?page=%v&pp=%v%v">%v</a></li>`,
			attr, i, sp.PerPage, query, i)
	}
	if sp.TotalPages > loogEnd {
		str += `<li><a>...</a></li>`
		str += fmt.Sprintf(`<li><a href="?page=%v&pp=%v%v">%v</a></li>`,
			sp.TotalPages, sp.PerPage, query, sp.TotalPages)
	}

	prev := sp.Page - 1
	next := sp.Page + 1
	prevQuery := fmt.Sprintf(`?page=%v&pp=%v%v`, prev, sp.PerPage, query)
	nextQuery := fmt.Sprintf(`?page=%v&pp=%v%v`, next, sp.PerPage, query)
	prevClass := ""
	nextClass := ""
	if next > sp.TotalPages {
		nextClass = "disabled"
		nextQuery = ""
	}
	if prev == 0 {
		prevClass = "disabled"
		prevQuery = ""
	}

	return template.HTML(`<nav aria-label="Page navigation" class="text-center">
	<ul class="pagination">
		<li class="` + prevClass + `">
			<a href="` + prevQuery + `" aria-label="Previous">
				<span aria-hidden="true">&laquo;</span>
			</a>
		</li>
` + str +
		`		<li class="` + nextClass + `">
			<a href="` + nextQuery + `" aria-label="Next">
				<span aria-hidden="true">&raquo;</span>
			</a>
		</li>
	</ul>
</nav>`)
}

func logoForHelper(name string) template.HTML {
	fontName := map[string]string{
		"gplus":    "google",
		"facebook": "facebook-official",
		"github":   "github",
	}
	return template.HTML(`<i class="fa fa-` + fontName[name] + `"></i>`)
}
