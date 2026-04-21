package dashboard

import (
	"net/http"

	"github.com/a-h/templ"
	"go_template/app/pages"
)

func DashboardForm(w http.ResponseWriter, r *http.Request) {
	templ.Handler(pages.UserForm()).ServeHTTP(w, r)
}
