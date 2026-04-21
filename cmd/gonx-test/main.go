package main

import (
	"fmt"
	"go_template/pkg/gonx"
	"os"
)

func main() {
	pf, err := gonx.ParseFile("app/pages/hello.gonx")
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	
	fmt.Printf("Package: %s\n", pf.Package)
	fmt.Printf("Funcs: %+v\n", pf.Funcs)
	fmt.Printf("Template:\n%s\n", pf.Template)
	
	compiler := gonx.NewCompiler(pf)
	code, err := compiler.Compile()
	if err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}
	
	fmt.Println("\n=== GENERATED CODE ===")
	fmt.Println(code)
	
	// Write file
	outPath := pf.OutputPath()
	if err := os.WriteFile(outPath, []byte(code), 0644); err != nil {
		fmt.Printf("Write error: %v\n", err)
		return
	}
	fmt.Printf("Written to: %s\n", outPath)
}
