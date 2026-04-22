package gonx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".." {
				return filepath.SkipDir
			}
			if info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		return nil
	})
	return found
}

// Compile procura todos os arquivos .gonx e os compila
func Compile(root string, verbose bool) error {
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

	for i, f := range files {
		start := time.Now()
		size, err := CompileFile(f)
		if err != nil {
			return fmt.Errorf("gonx %s: %w", f, err)
		}
		
		if verbose {
			if i < 10 {
				rel, _ := filepath.Rel(root, f)
				fmt.Printf("  compiled %s in %v (%d bytes)\n", rel, time.Since(start), size)
			} else if i == 10 {
				fmt.Println("  ...")
			}
		}
	}

	return nil
}

// CompileFile compila um único arquivo .gonx e retorna o tamanho do arquivo gerado
func CompileFile(path string) (int64, error) {
	pf, err := ParseFile(path)
	if err != nil {
		return 0, err
	}

	compiler := NewCompiler(pf)
	code, err := compiler.Compile()
	if err != nil {
		return 0, err
	}

	outPath := pf.OutputPath()
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return 0, err
	}
	if err := os.WriteFile(outPath, []byte(code), 0644); err != nil {
		return 0, err
	}

	info, err := os.Stat(outPath)
	if err != nil {
		return 0, nil
	}
	return info.Size(), nil
}
