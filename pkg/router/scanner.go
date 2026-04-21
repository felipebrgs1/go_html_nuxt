package router

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go_template/pkg/gonx"
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
	appDir := filepath.Join(s.root, "app")
	if _, err := os.Stat(appDir); err != nil {
		return nil, nil
	}

	return s.scanRecursive(appDir)
}

func (s *Scanner) scanRecursive(root string) ([]Route, error) {
	var routes []Route

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Pula diretórios internos que não são rotas
			name := info.Name()
			if name == "layouts" || name == "models" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Determina se é Página ou API baseado na extensão e localização
		isGonx := strings.HasSuffix(path, ".gonx")
		isGo := strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")

		if !isGonx && !isGo {
			return nil
		}

		// Import path base para o pacote
		// O prefixo depende se é código gerado (.gonx -> gonx/app/...) ou original (.go -> app/...)
		prefix := "app"
		if isGonx {
			prefix = "gonx/app"
		}

		pkgImport := filepath.Join(s.moduleName(), prefix, filepath.Dir(rel))
		if filepath.Dir(rel) == "." {
			pkgImport = filepath.Join(s.moduleName(), prefix)
		}

		if isGonx {
			pf, err := gonx.ParseFile(path)
			if err != nil {
				return fmt.Errorf("erro ao parsear %s: %w", path, err)
			}

			// .gonx são sempre páginas
			routes = append(routes, Route{
				Method:      "GET",
				Pattern:     fileToRoutePath(rel, true),
				FilePath:    path,
				PackagePath: pkgImport,
				HandlerName: pf.PageName,
				IsPage:      true,
				IsAPI:       false,
			})
		} else if isGo {
			handlers, err := s.parseHandlers(path)
			if err != nil {
				return fmt.Errorf("erro ao parsear %s: %w", path, err)
			}

			for _, h := range handlers {
				pattern := fileToRoutePath(rel, false)
				handlerSuffix := cleanHandlerName(h.name)
				if handlerSuffix != "" {
					pattern = filepath.Join(pattern, handlerSuffix)
				}
				
				routes = append(routes, Route{
					Method:      h.method,
					Pattern:     filepath.ToSlash(pattern),
					FilePath:    path,
					PackagePath: pkgImport,
					HandlerName: h.name,
					IsPage:      false,
					IsAPI:       true,
				})
			}
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
	// Fiber handler: func(c *fiber.Ctx) error
	if len(fn.Type.Params.List) == 1 {
		return true // Simplificação: assume que 1 param e retorno error é fiber
	}

	// Legado net/http: func(w, r)
	if len(fn.Type.Params.List) == 2 {
		return true
	}
	return false
}

func isGonxHandler(fn gonx.FuncSignature) bool {
	// Verifica se é handler HTTP: tem http.ResponseWriter e *http.Request OU *fiber.Ctx
	return strings.Contains(fn.Params, "http.ResponseWriter") || strings.Contains(fn.Params, "*fiber.Ctx")
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

func cleanHandlerName(name string) string {
	name = strings.TrimPrefix(name, "Get")
	name = strings.TrimPrefix(name, "Post")
	name = strings.TrimPrefix(name, "Put")
	name = strings.TrimPrefix(name, "Delete")
	name = strings.TrimPrefix(name, "Patch")
	return strings.ToLower(name)
}

func fileToRoutePath(rel string, isPage bool) string {
	// Remove extensão .go ou .gonx
	base := strings.TrimSuffix(rel, ".go")
	base = strings.TrimSuffix(base, ".gonx")
	// Converte separadores de path para /
	base = filepath.ToSlash(base)

	// api.go na raiz ou em subpastas
	if base == "api" || strings.HasSuffix(base, "/api") {
		base = strings.TrimSuffix(base, "api")
		base = strings.TrimSuffix(base, "/")
		if base == "" {
			return "/api"
		}
		// _id -> {id}
		parts := strings.Split(base, "/")
		for i, p := range parts {
			if strings.HasPrefix(p, "_") {
				parts[i] = ":" + p[1:]
			}
		}
		base = strings.Join(parts, "/")
		return "/api/" + base
	}

	// index.gonx -> /
	if base == "index" || strings.HasSuffix(base, "/index") {
		base = strings.TrimSuffix(base, "index")
		base = strings.TrimSuffix(base, "/")
		if base == "" {
			return "/"
		}
		// _id -> {id}
		parts := strings.Split(base, "/")
		for i, p := range parts {
			if strings.HasPrefix(p, "_") {
				parts[i] = ":" + p[1:]
			}
		}
		base = strings.Join(parts, "/")
		return "/" + base
	}

	// _id -> {id}
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
	// Outros arquivos .go que não se chamam api.go também são mapeados como /api/...
	return "/api/" + base
}
