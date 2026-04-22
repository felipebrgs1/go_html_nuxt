//go:build ignore

package main

import (
	"fmt"
	"os"

	"go_template/pkg/generator"
	"go_template/pkg/gonx"
	"go_template/pkg/router"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fmt.Printf("[generate] Compiling .gonx in %s...\n", path)
	if err := gonx.Compile(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling gonx: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[generate] Generating routes...")
	s := router.NewScanner(path)
	routes, err := s.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning routes: %v\n", err)
		os.Exit(1)
	}
	for _, r := range routes {
		fmt.Printf("  %s %s -> %s.%s\n", r.Method, r.Pattern, r.PkgImport, r.HandlerName)
	}
	if err := generator.Generate(path, routes); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating routes: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[generate] OK")
}
