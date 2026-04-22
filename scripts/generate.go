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
	fmt.Println("[generate] Compilando .gonx...")
	if err := gonx.Compile("."); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao compilar gonx: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[generate] Gerando rotas...")
	s := router.NewScanner(".")
	routes, err := s.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao scanear rotas: %v\n", err)
		os.Exit(1)
	}
	for _, r := range routes {
		fmt.Printf("  %s %s -> %s.%s\n", r.Method, r.Pattern, r.PkgImport, r.HandlerName)
	}
	if err := generator.Generate(".", routes); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gerar rotas: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[generate] OK")
}
