package router

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Route representa uma rota descoberta no sistema de arquivos
type Route struct {
	Method      string // GET, POST, PUT, DELETE, PATCH
	Pattern     string // ex: /users/{id}
	FilePath    string // caminho absoluto do arquivo
	PackagePath string // import path do pacote
	HandlerName string // nome da função handler
	IsPage      bool   // true se vem de app/pages
	IsAPI       bool   // true se vem de app/api
}

// Scanner varre app/pages e app/api procurando handlers
type Scanner struct {
	root string
}

func NewScanner(root string) *Scanner {
	return &Scanner{root: root}
}

func (s *Scanner) Scan() ([]Route, error) {
	var routes []Route

	pagesDir := filepath.Join(s.root, "app", "pages")
	apiDir := filepath.Join(s.root, "app", "api")

	if _, err := os.Stat(pagesDir); err == nil {
		rs, err := s.scanDir(pagesDir, "app/pages", true, false)
		if err != nil {
			return nil, err
		}
		routes = append(routes, rs...)
	}

	if _, err := os.Stat(apiDir); err == nil {
		rs, err := s.scanDir(apiDir, "app/api", false, true)
		if err != nil {
			return nil, err
		}
		routes = append(routes, rs...)
	}

	return routes, nil
}

func (s *Scanner) scanDir(dir, prefix string, isPage, isAPI bool) ([]Route, error) {
	var routes []Route

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		routePath := fileToRoutePath(rel, isPage)
		handlers, err := s.parseHandlers(path)
		if err != nil {
			return fmt.Errorf("erro ao parsear %s: %w", path, err)
		}

		pkgImport := filepath.Join(s.moduleName(), prefix, filepath.Dir(rel))
		if filepath.Dir(rel) == "." {
			pkgImport = filepath.Join(s.moduleName(), prefix)
		}

		for _, h := range handlers {
			routes = append(routes, Route{
				Method:      h.method,
				Pattern:     routePath,
				FilePath:    path,
				PackagePath: pkgImport,
				HandlerName: h.name,
				IsPage:      isPage,
				IsAPI:       isAPI,
			})
		}

		return nil
	})

	return routes, err
}

func (s *Scanner) moduleName() string {
	// Lê o module name do go.mod
	data, err := os.ReadFile(filepath.Join(s.root, "go.mod"))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

type handlerInfo struct {
	name   string
	method string
}

func (s *Scanner) parseHandlers(filePath string) ([]handlerInfo, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, src, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var handlers []handlerInfo
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}

		if !isHandlerFunc(fn) {
			continue
		}

		method := methodFromName(fn.Name.Name)
		handlers = append(handlers, handlerInfo{
			name:   fn.Name.Name,
			method: method,
		})
	}

	return handlers, nil
}

func isHandlerFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Params.List == nil {
		return false
	}
	if len(fn.Type.Params.List) != 2 {
		return false
	}

	// Verifica se o primeiro param é http.ResponseWriter e o segundo *http.Request
	// Para simplificar, apenas verificamos a quantidade de params e o retorno void
	// Uma verificação completa de tipos seria mais complexa com ast.
	// Aqui assumimos que qualquer função com 2 params e retorno vazio é um handler.
	return fn.Type.Results == nil || len(fn.Type.Results.List) == 0
}

func methodFromName(name string) string {
	switch {
	case strings.HasPrefix(name, "Get"):
		return "GET"
	case strings.HasPrefix(name, "Post"):
		return "POST"
	case strings.HasPrefix(name, "Put"):
		return "PUT"
	case strings.HasPrefix(name, "Delete"):
		return "DELETE"
	case strings.HasPrefix(name, "Patch"):
		return "PATCH"
	default:
		return "GET"
	}
}

func fileToRoutePath(rel string, isPage bool) string {
	// Remove extensão .go
	base := strings.TrimSuffix(rel, ".go")
	// Converte separadores de path para /
	base = filepath.ToSlash(base)

	// index.go -> /
	if base == "index" || strings.HasSuffix(base, "/index") {
		base = strings.TrimSuffix(base, "index")
		base = strings.TrimSuffix(base, "/")
		if base == "" {
			return "/"
		}
		// _id ou [id] -> {id}
		parts := strings.Split(base, "/")
		for i, p := range parts {
			if strings.HasPrefix(p, "_") {
				parts[i] = "{" + p[1:] + "}"
			}
		}
		base = strings.Join(parts, "/")
		return "/" + base
	}

	// _id ou [id] -> {id}
	parts := strings.Split(base, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "_") {
			parts[i] = "{" + p[1:] + "}"
		}
	}
	base = strings.Join(parts, "/")

	if isPage {
		return "/" + base
	}
	// API routes prefixam com /api
	return "/api/" + base
}
