package backEnd

import (
	"net/http"
	"text/template"
)

var (
	tpl = template.Must(template.ParseGlob("./frontEnd/*html"))
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tpl.ExecuteTemplate(w, "index.html", nil)

}
