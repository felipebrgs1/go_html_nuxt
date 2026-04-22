package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go_template/pkg/gonx"
	"go_template/pkg/router"
)

// Severity levels
type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
)

// Issue represents a lint finding
type Issue struct {
	File     string
	Line     int
	Rule     string
	Severity Severity
	Message  string
}

// Result aggregates all issues
type Result struct {
	Issues []Issue
}

func (r *Result) Add(file string, line int, rule string, sev Severity, msg string) {
	r.Issues = append(r.Issues, Issue{
		File:     file,
		Line:     line,
		Rule:     rule,
		Severity: sev,
		Message:  msg,
	})
}

func (r *Result) HasErrors() bool {
	for _, i := range r.Issues {
		if i.Severity == Error {
			return true
		}
	}
	return false
}

func (r *Result) Sort() {
	sort.Slice(r.Issues, func(i, j int) bool {
		if r.Issues[i].File != r.Issues[j].File {
			return r.Issues[i].File < r.Issues[j].File
		}
		return r.Issues[i].Line < r.Issues[j].Line
	})
}

// Linter runs all rules
type Linter struct {
	root   string
	result *Result
}

func New(root string) *Linter {
	return &Linter{root: root, result: &Result{}}
}

func (l *Linter) Run() (*Result, error) {
	// Rule: no orphan templ files
	l.checkOrphanTempl()

	// Rule: gonx syntax and semantics
	l.checkGonxFiles()

	// Rule: route duplicates and conflicts
	l.checkRoutes()

	// Rule: htmx attributes in templates
	l.checkHtmx()

	l.result.Sort()
	return l.result, nil
}

// --------------------------------------------------------------------------
// Rule: orphan templ / _templ.go files
// --------------------------------------------------------------------------

func (l *Linter) checkOrphanTempl() {
	appDir := filepath.Join(l.root, "app")
	_ = filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if !strings.HasSuffix(name, ".templ") && !strings.HasSuffix(name, "_templ.go") {
			return nil
		}
		base := strings.TrimSuffix(name, ".templ")
		base = strings.TrimSuffix(base, "_templ.go")
		dir := filepath.Dir(path)
		gonxPath := filepath.Join(dir, base+".gonx")
		if _, err := os.Stat(gonxPath); err == nil {
			l.result.Add(path, 1, "no-orphan-templ", Warning,
				fmt.Sprintf("arquivo %s existe junto com %s.gonx; considere remover o arquivo obsoleto", name, base))
		}
		return nil
	})
}

// --------------------------------------------------------------------------
// Rule: gonx file validation
// --------------------------------------------------------------------------

func (l *Linter) checkGonxFiles() {
	appDir := filepath.Join(l.root, "app")
	_ = filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".gonx") {
			return nil
		}
		l.lintGonx(path)
		return nil
	})
}

func (l *Linter) lintGonx(path string) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		l.result.Add(path, 1, "gonx-read", Error, err.Error())
		return
	}
	content := string(contentBytes)

	pf, err := gonx.ParseFile(path)
	if err != nil {
		l.result.Add(path, 1, "gonx-parse", Error, fmt.Sprintf("falha ao parsear: %v", err))
		return
	}

	// Must have template block
	if pf.Template == "" {
		l.result.Add(path, 1, "gonx-blocks", Error, "bloco <template> vazio ou ausente")
	}

	// Must have script block
	if pf.Script == "" {
		l.result.Add(path, 1, "gonx-blocks", Error, "bloco <script> vazio ou ausente")
	}

	// Handler signature validation
	for _, fn := range pf.Funcs {
		if !isValidHandler(fn) {
			// Try to find line number
			line := findLine(content, "func "+fn.Name)
			l.result.Add(path, line, "handler-signature", Error,
				fmt.Sprintf("função %s não parece ser um handler Fiber válido (esperado: func %s(c *fiber.Ctx) error)", fn.Name, fn.Name))
		}
	}

	// HTML balance check (simple stack)
	if pf.Template != "" {
		line, msg := checkHTMLBalance(pf.Template)
		if msg != "" {
			l.result.Add(path, line, "html-balanced", Warning, msg)
		}
	}
}

func isValidHandler(fn gonx.FuncSignature) bool {
	// Aceita assinaturas Fiber (error return) ou legado io.Writer (para layouts/templates internos)
	if strings.Contains(fn.Params, "*fiber.Ctx") {
		return fn.ReturnType == "error" || fn.ReturnType == ""
	}
	// Legado net/http
	return strings.Contains(fn.Params, "http.ResponseWriter") && strings.Contains(fn.Params, "*http.Request")
}

// --------------------------------------------------------------------------
// Rule: route duplicates
// --------------------------------------------------------------------------

func (l *Linter) checkRoutes() {
	scanner := router.NewScanner(l.root)
	routes, err := scanner.Scan()
	if err != nil {
		l.result.Add("routes.gen.go", 1, "route-scan", Error, fmt.Sprintf("falha ao scanear rotas: %v", err))
		return
	}

	keyed := make(map[string][]router.Route)
	for _, r := range routes {
		key := r.Method + " " + r.Pattern
		keyed[key] = append(keyed[key], r)
	}

	for key, rs := range keyed {
		if len(rs) > 1 {
			var handlers []string
			for _, r := range rs {
				handlers = append(handlers, fmt.Sprintf("%s.%s", r.PkgImport, r.HandlerName))
			}
			l.result.Add("routes.gen.go", 1, "route-duplicate", Error,
				fmt.Sprintf("rota duplicada %s registrada por: %s", key, strings.Join(handlers, ", ")))
		}
	}
}

// --------------------------------------------------------------------------
// Rule: htmx attribute hygiene
// --------------------------------------------------------------------------

var htmxTriggerRe = regexp.MustCompile(`\bhx-(get|post|put|delete|patch)\s*=\s*"([^"]+)"`)
var htmxSwapRe = regexp.MustCompile(`\bhx-swap\s*=\s*"([^"]+)"`)
var htmxTargetRe = regexp.MustCompile(`\bhx-target\s*=\s*"([^"]+)"`)

func (l *Linter) checkHtmx() {
	appDir := filepath.Join(l.root, "app")
	_ = filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".gonx") {
			return nil
		}
		l.lintHtmxInFile(path)
		return nil
	})
}

func (l *Linter) lintHtmxInFile(path string) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(contentBytes)

	template := gonx.ExtractBlock(content, "template")
	lines := strings.Split(template, "\n")

	for i, line := range lines {
		lineNum := i + 1
		if !htmxTriggerRe.MatchString(line) {
			continue
		}

		swapMatch := htmxSwapRe.FindStringSubmatch(line)
		targetMatch := htmxTargetRe.FindStringSubmatch(line)

		// If hx-swap="delete" and no hx-target, that's usually wrong
		if len(swapMatch) > 1 && swapMatch[1] == "delete" && len(targetMatch) == 0 {
			l.result.Add(path, lineNum, "htmx-target", Warning,
				"hx-swap=\"delete\" sem hx-target pode falhar silenciosamente; adicione hx-target=\"closest tr\" ou similar")
		}

		// If hx-swap="outerHTML" on a PUT/POST without target, warn
		if len(swapMatch) > 1 && swapMatch[1] == "outerHTML" && len(targetMatch) == 0 {
			l.result.Add(path, lineNum, "htmx-target", Warning,
				"hx-swap=\"outerHTML\" sem hx-target pode substituir o elemento errado")
		}

		// If trigger is present but no target at all (and no swap=none)
		if len(targetMatch) == 0 && (len(swapMatch) == 0 || swapMatch[1] != "none") {
			l.result.Add(path, lineNum, "htmx-target", Info,
				"requisição HTMX sem hx-target; o default é o próprio elemento (pode ser intencional)")
		}
	}
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func findLine(content, substr string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i + 1
		}
	}
	return 1
}

func checkHTMLBalance(template string) (int, string) {
	// Very simplistic HTML tag balancer: checks <tag> vs </tag> for common tags
	// Ignores self-closing tags and void elements
	re := regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)[^>]*?>`)
	type stackItem struct {
		tag  string
		line int
	}
	var stack []stackItem

	lines := strings.Split(template, "\n")
	for lineIdx, line := range lines {
		matches := re.FindAllStringSubmatchIndex(line, -1)
		for _, m := range matches {
			slash := line[m[2]:m[3]]
			tag := strings.ToLower(line[m[4]:m[5]])
			if isVoidTag(tag) {
				continue
			}
			if slash == "/" {
				if len(stack) == 0 {
					return lineIdx + 1, fmt.Sprintf("tag de fechamento </%s> sem tag de abertura correspondente", tag)
				}
				top := stack[len(stack)-1]
				if top.tag != tag {
					return lineIdx + 1, fmt.Sprintf("tag de fechamento </%s> não corresponde à tag de abertura <%s> (linha %d)", tag, top.tag, top.line)
				}
				stack = stack[:len(stack)-1]
			} else {
				stack = append(stack, stackItem{tag: tag, line: lineIdx + 1})
			}
		}
	}
	if len(stack) > 0 {
		top := stack[len(stack)-1]
		return top.line, fmt.Sprintf("tag <%s> não foi fechada", top.tag)
	}
	return 0, ""
}

func isVoidTag(tag string) bool {
	switch tag {
	case "br", "hr", "img", "input", "meta", "link", "area", "base", "col", "embed",
		"param", "source", "track", "wbr", "!doctype", "svg", "path", "circle", "rect",
		"line", "polyline", "polygon", "ellipse", "g", "defs", "use", "text", "tspan":
		return true
	}
	return false
}
