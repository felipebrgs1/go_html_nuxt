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

// Route representa uma rota mapeada
type Route struct {
	Method      string
	Pattern     string
	PkgImport    string
	HandlerName string
	IsPage      bool
	SourceFile  string
}

// Scanner percorre o diretório app/ para encontrar rotas
type Scanner struct {
	root string
}

func NewScanner(root string) *Scanner {
	return &Scanner{root: root}
}

func (s *Scanner) Scan() ([]Route, error) {
	appDir := filepath.Join(s.root, "app")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("diretório app/ não encontrado")
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
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		isGonx := strings.HasSuffix(path, ".gonx")
		isGo := strings.HasSuffix(path, ".go")

		if !isGonx && !isGo {
			return nil
		}

		// Ignora layouts e components no escaneamento de rotas
		if strings.Contains(rel, "layouts/") || strings.Contains(rel, "components/") || strings.Contains(rel, "models/") {
			return nil
		}

		if isGonx {
			pf, err := gonx.ParseFile(path)
			if err != nil {
				return nil // Pula arquivos inválidos
			}

			// Define o prefixo de importação correto (gonx/app/...)
			prefix := "gonx/app"
			pkgImport := filepath.Join(s.moduleName(), prefix, filepath.Dir(rel))

			// 1. Adiciona rotas extras encontradas no script (Get, Post, etc) PRIMEIRO
			// Isso garante que se houver um Get(), ele tenha precedência sobre o Index() padrão
			for _, fn := range pf.Funcs {
				method := getMethod(fn.Name)
				if method == "" {
					continue
				}

				pattern := s.fileToRoutePath(rel, true)
				handlerSuffix := cleanHandlerName(fn.Name)
				if handlerSuffix != "" {
					pattern = filepath.Join(pattern, handlerSuffix)
				}
				
				// Garante que o pattern comece com / e use /
				pattern = filepath.ToSlash(pattern)
				if !strings.HasPrefix(pattern, "/") {
					pattern = "/" + pattern
				}

				routes = append(routes, Route{
					Method:      method,
					Pattern:     pattern,
					PkgImport:    pkgImport,
					HandlerName: fn.Name,
					IsPage:      false,
					SourceFile:  path,
				})
			}

			// 2. Adiciona a rota principal de renderização da página (apenas se não houver um Get() conflitando)
			hasGetConflict := false
			mainPattern := s.fileToRoutePath(rel, true)
			for _, r := range routes {
				if r.SourceFile == path && r.Method == "GET" && r.Pattern == mainPattern && r.HandlerName != pf.PageName {
					hasGetConflict = true
					break
				}
			}

			if !hasGetConflict {
				routes = append(routes, Route{
					Method:      "GET",
					Pattern:     mainPattern,
					PkgImport:    pkgImport,
					HandlerName: pf.PageName,
					IsPage:      true,
					SourceFile:  path,
				})
			}
		} else if isGo {
			// Analisa arquivos .go para encontrar handlers (GetX, PostX, etc)
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return nil
			}

			pkgImport := filepath.Join(s.moduleName(), "app", filepath.Dir(rel))

			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil {
					continue
				}

				method := getMethod(fn.Name.Name)
				if method == "" {
					continue
				}

				pattern := s.fileToRoutePath(rel, false)
				handlerSuffix := cleanHandlerName(fn.Name.Name)
				if handlerSuffix != "" {
					pattern = filepath.Join(pattern, handlerSuffix)
				}
				
				// Garante que o pattern comece com / e use /
				pattern = filepath.ToSlash(pattern)
				if !strings.HasPrefix(pattern, "/") {
					pattern = "/" + pattern
				}

				routes = append(routes, Route{
					Method:      method,
					Pattern:     pattern,
					PkgImport:    pkgImport,
					HandlerName: fn.Name.Name,
					IsPage:      false,
					SourceFile:  path,
				})
			}
		}

		return nil
	})

	return routes, err
}

func (s *Scanner) moduleName() string {
	// Simplificado para o template
	return "go_template"
}

func getMethod(name string) string {
	if strings.HasPrefix(name, "Get") {
		return "GET"
	}
	if strings.HasPrefix(name, "Post") {
		return "POST"
	}
	if strings.HasPrefix(name, "Put") {
		return "PUT"
	}
	if strings.HasPrefix(name, "Delete") {
		return "DELETE"
	}
	if strings.HasPrefix(name, "Patch") {
		return "PATCH"
	}
	return ""
}

func cleanHandlerName(name string) string {
	prefixes := []string{"Get", "Post", "Put", "Delete", "Patch"}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			name = strings.TrimPrefix(name, p)
			break
		}
	}
	if name == "" {
		return ""
	}
	// Converte para lowercase
	return strings.ToLower(name)
}

func (s *Scanner) fileToRoutePath(rel string, isPage bool) string {
	base := strings.TrimSuffix(rel, ".go")
	base = strings.TrimSuffix(base, ".gonx")
	base = filepath.ToSlash(base)

	if isPage {
		// Remove prefixo pages/ se existir
		if after, ok :=strings.CutPrefix(base, "pages/"); ok  {
			base = after
		}
		// index -> /
		if base == "index" || strings.HasSuffix(base, "/index") {
			base = strings.TrimSuffix(base, "index")
			base = strings.TrimSuffix(base, "/")
		}
		
		return s.applyParams(base, true)
	}

	// API Handlers (.go)
	if strings.HasPrefix(base, "api/") {
		base = strings.TrimPrefix(base, "api/")
	} else if strings.HasPrefix(base, "pages/") {
		base = strings.TrimPrefix(base, "pages/")
	}

	// Remove "api" ou "index" do final
	if base == "api" || strings.HasSuffix(base, "/api") || base == "index" || strings.HasSuffix(base, "/index") {
		base = strings.TrimSuffix(base, "api")
		base = strings.TrimSuffix(base, "index")
		base = strings.TrimSuffix(base, "/")
	}

	if base == "" {
		return "/api"
	}
	return s.applyParams("/api/"+base, false)
}

func (s *Scanner) applyParams(base string, leadingSlash bool) string {
	parts := strings.Split(base, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "_") {
			parts[i] = ":" + p[1:]
		}
	}
	res := strings.Join(parts, "/")
	if leadingSlash && !strings.HasPrefix(res, "/") {
		res = "/" + res
	}
	return res
}
