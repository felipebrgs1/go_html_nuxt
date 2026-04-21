package api

import (
	"net/http"

	"github.com/a-h/templ"
	"go_template/app/pages"
	"go_template/pkg/htmx"
)

func GetUsersSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var filtered []pages.User
	for _, u := range pages.Users {
		if query == "" || contains(u.Name, query) || contains(u.Email, query) {
			filtered = append(filtered, u)
		}
	}
	templ.Handler(pages.UserTable(filtered)).ServeHTTP(w, r)
}

func PostUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	role := r.FormValue("role")

	newUser := pages.User{
		ID:     len(pages.Users) + 1,
		Name:   name,
		Email:  email,
		Role:   role,
		Status: "active",
		Avatar: initials(name),
	}
	pages.Users = append(pages.Users, newUser)

	if htmx.IsHTMXRequest(r) {
		htmx.Trigger(w, "userCreated")
		templ.Handler(pages.UserRow(newUser)).ServeHTTP(w, r)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func initials(name string) string {
	parts := []rune(name)
	if len(parts) >= 2 {
		return string(parts[0]) + string(parts[len(parts)-1])
	}
	return string(parts[0])
}
