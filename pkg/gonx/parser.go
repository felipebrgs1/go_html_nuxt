package gonx

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParsedFile representa um arquivo .gonx parseado
type ParsedFile struct {
	Package     string
	Imports     []string         // imports explícitos do usuário
	AutoImports []string         // imports descobertos automaticamente
	Script      string           // código Go puro do bloco <script>
	Template    string           // conteúdo do bloco <template>
	Style       string           // conteúdo do bloco <style>
	Funcs       []FuncSignature  // funções encontradas no script
	FilePath    string
	Root        string           // raiz do projeto (para resolver auto-imports)
	IsPage      bool             // true se vem de app/pages
	IsAPI       bool             // true se vem de app/api
	PageName    string           // nome do handler para páginas (PascalCase do arquivo)
}

// FuncSignature representa uma função encontrada no script
type FuncSignature struct {
	Name       string
	Params     string
	ReturnType string
	LayoutPkg  string // pacote do layout (ex: "layouts")
	LayoutFunc string // função do layout (ex: "Base")
	LayoutArgs string // argumentos extra do layout (ex: `"Dashboard"`)
	HasBody    bool   // true se a função tem corpo no script
	BodySource string // texto completo da função no script
}

func (pf *ParsedFile) ModuleName() string {
	data, err := os.ReadFile(filepath.Join(pf.Root, "go.mod"))
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
func ParseFile(path string) (*ParsedFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	root := findProjectRoot(path)
	pf := &ParsedFile{FilePath: path, Root: root}

	// Detecta se é página ou API
	pf.IsPage = filepath.Ext(path) == ".gonx"
	pf.IsAPI = !pf.IsPage
	pf.PageName = fileToHandlerName(path)

	// Extrai blocos <template>, <script>, <style>
	pf.Template = ExtractBlock(string(content), "template")
	pf.Script = ExtractBlock(string(content), "script")
	pf.Style = ExtractBlock(string(content), "style")

	// Parse o script para extrair package, imports e funções
	if err := pf.parseScript(); err != nil {
		return nil, fmt.Errorf("erro ao parsear script: %w", err)
	}

	// Auto-import: descobre pacotes internos usados
	if err := pf.resolveAutoImports(); err != nil {
		return nil, fmt.Errorf("erro ao resolver auto-imports: %w", err)
	}

	return pf, nil
}

func fileToHandlerName(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	// Converte para PascalCase (ex: hello-world -> HelloWorld, index -> Index)
	parts := strings.Split(base, "-")
	var result string
	for _, p := range parts {
		if p == "" {
			continue
		}
		result += strings.ToUpper(p[:1]) + p[1:]
	}
	return result
}

func findProjectRoot(filePath string) string {
	dir := filepath.Dir(filePath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dir
}

func ExtractBlock(content, tag string) string {
	re := regexp.MustCompile(`(?s)<` + tag + `[^>]*>(.*?)</` + tag + `>`)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return ""
	}
	// Pega o match com maior conteúdo para evitar tags vazias no template
	best := ""
	for _, m := range matches {
		if len(m) >= 2 && len(strings.TrimSpace(m[1])) > len(best) {
			best = strings.TrimSpace(m[1])
		}
	}
	return best
}

func (pf *ParsedFile) parseScript() error {
	if pf.Script == "" {
		return fmt.Errorf("bloco <script> não encontrado")
	}

	originalSrc := pf.Script
	src := originalSrc
	offset := 0
	if !strings.HasPrefix(strings.TrimSpace(src), "package ") {
		prefix := "package main\n"
		src = prefix + src
		offset = len(prefix)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "script.go", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}

	pf.Package = f.Name.Name

	// Extrai imports explícitos
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil {
			pf.Imports = append(pf.Imports, imp.Name.Name+` "`+path+`"`)
		} else {
			pf.Imports = append(pf.Imports, `"`+path+`"`)
		}
	}

	// Extrai funções
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}

		sig := FuncSignature{Name: fn.Name.Name}
		if fn.Type.Params != nil {
			var params []string
			for _, field := range fn.Type.Params.List {
				for _, name := range field.Names {
					params = append(params, name.Name+" "+exprToString(field.Type))
				}
			}
			sig.Params = strings.Join(params, ", ")
		}
		if fn.Type.Results != nil {
			var rets []string
			for _, field := range fn.Type.Results.List {
				rets = append(rets, exprToString(field.Type))
			}
			sig.ReturnType = strings.Join(rets, ", ")
		}

		// Extrai corpo da função
		if fn.Body != nil {
			sig.HasBody = len(fn.Body.List) > 0
			start := fset.Position(fn.Pos()).Offset - offset
			end := fset.Position(fn.End()).Offset - offset
			if start >= 0 && end > start && end <= len(originalSrc) {
				sig.BodySource = originalSrc[start:end]
			}
		}

		// Extrai diretivas de layout dos comentários
		if fn.Doc != nil {
			for _, c := range fn.Doc.List {
				text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
				if strings.HasPrefix(text, "gonx:layout ") {
					layout := strings.TrimSpace(strings.TrimPrefix(text, "gonx:layout "))
					dotIdx := strings.LastIndex(layout, ".")
					if dotIdx > 0 {
						sig.LayoutPkg = layout[:dotIdx]
						sig.LayoutFunc = layout[dotIdx+1:]
					}
				}
				if strings.HasPrefix(text, "gonx:layout-args ") {
					sig.LayoutArgs = strings.TrimSpace(strings.TrimPrefix(text, "gonx:layout-args "))
				}
			}
		}

		pf.Funcs = append(pf.Funcs, sig)
	}

	return nil
}

// resolveAutoImports scaneia o projeto e adiciona imports para pacotes internos usados
func (pf *ParsedFile) resolveAutoImports() error {
	if pf.Root == "" {
		return nil
	}

	// Coleta todos os identificadores usados no script
	idents := pf.extractIdentifiers()
	if len(idents) == 0 {
		return nil
	}

	// Mapeia pacotes disponíveis em app/ e pkg/
	pkgMap, err := pf.scanProjectPackages()
	if err != nil {
		return err
	}

	// Imports já explícitos (para não duplicar)
	explicit := make(map[string]bool)
	explicitPkgName := make(map[string]bool)
	for _, imp := range pf.Imports {
		// extrai o path do import
		parts := strings.Split(imp, `"`)
		if len(parts) >= 2 {
			path := parts[1]
			explicit[path] = true
			// último segmento do path é o nome do pacote
			segs := strings.Split(path, "/")
			if len(segs) > 0 {
				explicitPkgName[segs[len(segs)-1]] = true
			}
		}
	}

	// Para cada identificador usado, verifica se é um pacote interno
	for ident := range idents {
		if explicitPkgName[ident] {
			continue
		}
		if pkgPath, ok := pkgMap[ident]; ok {
			if !explicit[pkgPath] {
				pf.AutoImports = append(pf.AutoImports, `"`+pkgPath+`"`)
				explicit[pkgPath] = true
			}
		}
	}

	return nil
}

// extractIdentifiers extrai todos os identificadores de pacote usados no script
func (pf *ParsedFile) extractIdentifiers() map[string]bool {
	idents := make(map[string]bool)
	
	src := pf.Script
	if !strings.HasPrefix(strings.TrimSpace(src), "package ") {
		src = "package main\n" + src
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "script.go", src, parser.AllErrors)
	if err != nil {
		return idents
	}

	// Percorre o AST procurando SelectorExpr (ex: layouts.Base)
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if ident, ok := x.X.(*ast.Ident); ok {
				idents[ident.Name] = true
			}
		case *ast.CallExpr:
			// Também detecta chamadas diretas de pacote (ex: api.GetHello())
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					idents[ident.Name] = true
				}
			}
		}
		return true
	})

	// Também adiciona layouts referenciados por diretivas
	for _, fn := range pf.Funcs {
		if fn.LayoutPkg != "" {
			idents[fn.LayoutPkg] = true
		}
	}

	return idents
}

// scanProjectPackages mapeia nomes de pacotes para seus import paths
func (pf *ParsedFile) scanProjectPackages() (map[string]string, error) {
	pkgMap := make(map[string]string)
	
	moduleName := pf.moduleName()
	if moduleName == "" {
		return pkgMap, nil
	}

	// Scaneia app/ e pkg/
	for _, prefix := range []string{"app", "pkg"} {
		dir := filepath.Join(pf.Root, prefix)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return nil
			}
			// Verifica se é um pacote Go ou Gonx
			entries, _ := os.ReadDir(path)
			hasGo := false
			hasGonx := false
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
					hasGo = true
				}
				if strings.HasSuffix(name, ".gonx") {
					hasGonx = true
				}
			}
			if !hasGo && !hasGonx {
				return nil
			}

			rel, _ := filepath.Rel(pf.Root, path)
			importPath := filepath.Join(moduleName, rel)
			importPath = filepath.ToSlash(importPath)
			pkgName := filepath.Base(path)
			
			// Se tem .gonx, registra o caminho compilado em gonx/ (prioridade)
			if hasGonx {
				gonxPath := filepath.Join(moduleName, "gonx", rel)
				gonxPath = filepath.ToSlash(gonxPath)
				if _, exists := pkgMap[pkgName]; !exists {
					pkgMap[pkgName] = gonxPath
				}
			}

			// Se tem .go tradicional, registra o caminho original
			if hasGo {
				if _, exists := pkgMap[pkgName]; !exists {
					pkgMap[pkgName] = importPath
				}
			}
			return nil
		})
	}

	return pkgMap, nil
}

func (pf *ParsedFile) moduleName() string {
	data, err := os.ReadFile(filepath.Join(pf.Root, "go.mod"))
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

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.FuncType:
		var params []string
		if e.Params != nil {
			for _, field := range e.Params.List {
				ptype := exprToString(field.Type)
				for range field.Names {
					params = append(params, ptype)
				}
				if len(field.Names) == 0 {
					params = append(params, ptype)
				}
			}
		}
		var results []string
		if e.Results != nil {
			for _, field := range e.Results.List {
				results = append(results, exprToString(field.Type))
			}
		}
		sig := "func(" + strings.Join(params, ", ") + ")"
		if len(results) == 1 {
			sig += " " + results[0]
		} else if len(results) > 1 {
			sig += " (" + strings.Join(results, ", ") + ")"
		}
		return sig
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		return "chan"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// OutputPath retorna o caminho do arquivo gerado dentro de .gonx/
func (pf *ParsedFile) OutputPath() string {
	rel, _ := filepath.Rel(pf.Root, pf.FilePath)
	base := filepath.Base(rel)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	dir := filepath.Dir(rel)
	return filepath.Join(pf.Root, "gonx", dir, name+".go")
}
