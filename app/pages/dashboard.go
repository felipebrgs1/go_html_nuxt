package pages

import (
	"net/http"

	"github.com/a-h/templ"
	"go_template/pkg/htmx"
)

type User struct {
	ID     int
	Name   string
	Email  string
	Role   string
	Status string
	Avatar string
}

var Users = []User{
	{1, "Ana Silva", "ana@exemplo.com", "Admin", "active", "AS"},
	{2, "Bruno Costa", "bruno@exemplo.com", "Editor", "active", "BC"},
	{3, "Carla Mendes", "carla@exemplo.com", "Viewer", "inactive", "CM"},
	{4, "Diego Souza", "diego@exemplo.com", "Editor", "active", "DS"},
	{5, "Elisa Prado", "elisa@exemplo.com", "Viewer", "active", "EP"},
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	if htmx.IsHTMXRequest(r) {
		templ.Handler(dashboardContent()).ServeHTTP(w, r)
		return
	}
	templ.Handler(DashboardPage()).ServeHTTP(w, r)
}


