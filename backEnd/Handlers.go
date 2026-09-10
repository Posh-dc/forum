package backEnd

import (
	"net/http"
	"text/template"
)

var (
	tpl = template.Must(template.ParseGlob("./static/*html"))
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tpl.ExecuteTemplate(w, "index.html", nil)

}
