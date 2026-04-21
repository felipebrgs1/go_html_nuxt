package gonx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HasGonx verifica se existe algum arquivo .gonx no projeto
func HasGonx(root string) bool {
	found := false
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".gonx" {
			found = true
			return filepath.SkipAll
		}
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

// Compile procura todos os arquivos .gonx e os compila
func Compile(root string) error {
	var files []string
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".." {
				return filepath.SkipDir
			}
			if info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".gonx" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := CompileFile(f); err != nil {
			return fmt.Errorf("gonx %s: %w", f, err)
		}
	}

	return nil
}

// CompileFile compila um único arquivo .gonx
func CompileFile(path string) error {
	pf, err := ParseFile(path)
	if err != nil {
		return err
	}

	compiler := NewCompiler(pf)
	code, err := compiler.Compile()
	if err != nil {
		return err
	}

	outPath := pf.OutputPath()
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(code), 0644)
}
