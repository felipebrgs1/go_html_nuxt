package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"go_template/.framework"
)

func main() {
	routeMux := http.NewServeMux()
	framework.RegisterRoutes(routeMux)

	fileServer := http.FileServer(http.Dir("public"))

	// Handler final: primeiro tenta arquivo estático, depois rotas geradas
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join("public", filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		routeMux.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
