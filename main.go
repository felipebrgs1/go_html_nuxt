package main

import (
	"fmt"
	"net/http"
	"os"

	"go_template/.framework"
)

func main() {
	mux := http.NewServeMux()
	framework.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
