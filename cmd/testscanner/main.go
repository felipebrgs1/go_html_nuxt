package main

import (
	"fmt"
	"go_template/pkg/router"
)

func main() {
	scanner := router.NewScanner(".")
	routes, err := scanner.Scan()
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	for _, r := range routes {
		fmt.Printf("%s %s -> %s.%s (%s)\n", r.Method, r.Pattern, r.PackagePath, r.HandlerName, r.FilePath)
	}
}
