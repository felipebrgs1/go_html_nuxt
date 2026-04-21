package pages

import (
	"fmt"
	"net/http"
)

func User(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintf(w, "<h1>User ID: %s</h1>", id)
}
