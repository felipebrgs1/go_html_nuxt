package pages

import (
	"net/http"

	"github.com/a-h/templ"
)

func Index(w http.ResponseWriter, r *http.Request) {
	templ.Handler(IndexPage()).ServeHTTP(w, r)
}
