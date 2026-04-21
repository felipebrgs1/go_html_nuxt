# Go Template Framework — Agent Guide

## Overview

Nuxt-like full-stack framework in **Go + HTMX + TailwindCSS**. Zero-config, hot-reload, file-based routing. Uses a custom template language (`.gonx`) that compiles to Go.

## Architecture

```
app/
  pages/          # File-based routes (pages)
    index.gonx    # GET /
    hello.gonx    # GET /hello
    dashboard.gonx # GET /dashboard
    users/
      _id.gonx    # GET /users/{id}
  api/            # API routes
    hello.go      # GET /api/hello
    users.go      # GET/POST/DELETE /api/users
  layouts/        # Shared layouts & components
    base.gonx     # HTML shell
    components.gonx # Reusable components (Card, etc.)

gonx/            # Generated Go code (auto-created, .gitignored)
  app/
    pages/
      index.go
      hello.go
framework_gen/
  routes.gen.go   # Auto-generated route registry

public/           # Static assets (CSS, images, JS)
  styles.css
```

## `.gonx` File Format

Single-file components (SFC) with 3 blocks:

```gonx
<template>
  <div class="p-4">
    <h1>{{ pageTitle }}</h1>
    {{ if isAdmin }}
      <span>Admin</span>
    {{ end }}
    <ul>
      {{ range users }}
        <li>{{ .Name }} — {{ .Email }}</li>
      {{ end }}
    </ul>
  </div>
</template>

<script>
package pages

import "net/http"

func Index(w http.ResponseWriter, r *http.Request) {
  // This becomes the HTTP handler
  // Auto-imports resolve packages automatically (layouts, api, etc.)
}
</script>

<style>
/* Optional — gets injected as-is */
</style>
```

### Template Syntax

| Feature | Syntax |
|---------|--------|
| Interpolation | `{{ variable }}` |
| Raw HTML | `{{ html.rawString }}` or `{{ html.SomeVar }}` |
| If/Else | `{{ if cond }} ... {{ else }} ... {{ end }}` |
| Range | `{{ range items }} ... {{ end }}` |
| Component call | `{{ call ComponentName arg1 arg2 }}` |

**HTML escaping:** All `{{ }}` output is HTML-escaped by default. Use `html.` prefix for raw output.

### Script Block

- Must declare `package <name>` (usually matches directory: `pages`, `api`, `layouts`)
- Any exported function matching `func Xxx(w http.ResponseWriter, r *http.Request)` becomes a route handler
- **Auto-imports:** The compiler scans your code and automatically adds imports for internal packages (`layouts`, `api`, `pkg/htmx`, etc.). You only need to explicitly import external packages.

## Routing

File-based routing, derived from file path + function name prefix:

| File | Handler | Route |
|------|---------|-------|
| `app/pages/index.gonx` | `func Index(...)` | `GET /` |
| `app/pages/hello.gonx` | `func Hello(...)` | `GET /hello` |
| `app/pages/users/_id.gonx` | `func User(...)` | `GET /users/{id}` |
| `app/api/hello.go` | `func GetHello(...)` | `GET /api/hello` |
| `app/api/users.go` | `func PostUsers(...)` | `POST /api/users` |

**Method prefixes:** `Get`, `Post`, `Put`, `Delete`, `Patch` → HTTP method. No prefix = `GET`.

## Components / Layouts

Components are just `.gonx` files with render functions. Call them from templates:

```gonx
{{ call Card "Recent Activity" }}
  <ul><li>Item 1</li></ul>
{{ endcall }}
```

Or from Go code:

```go
func MyPage(w http.ResponseWriter, r *http.Request) {
  layouts.RenderBase(w, "My Title", func(w io.Writer) {
    RenderMyPageContent(w)
  })
}
```

## HTMX Integration

HTMX 2.x is included in the base layout. Use directly in templates:

```html
<button hx-get="/api/hello" hx-target="#result" hx-swap="innerHTML">
  Load
</button>
<div id="result"></div>
```

Server-side, detect HTMX requests:

```go
import "go_template/pkg/htmx"

if htmx.IsHTMXRequest(r) {
  // Return partial HTML
} else {
  // Return full page
}
```

## Development

```bash
# Start dev server with hot-reload
make dev
# or
go run cmd/framework/main.go dev
```

Watches: `.go`, `.templ`, `.gonx`, `.css`

Build pipeline on file change:
1. Compile `.gonx` → `gonx/` directory
2. Compile `.templ` → `_templ.go` (legacy, being phased out)
3. Generate routes → `framework_gen/routes.gen.go`
4. Compile Tailwind → `public/styles.css`
5. Restart Go server

## Build for Production

```bash
go build -o bin/app .
```

The server uses `PORT` env var (default `3000`).

## Adding a New Page

1. Create `app/pages/mypage.gonx`
2. Add `<template>` with HTML
3. Add `<script>` with handler function:
   ```go
   func MyPage(w http.ResponseWriter, r *http.Request) {
     RenderMyPage(w)
   }
   ```
4. Server auto-restarts. Access at `GET /mypage`.

## Adding API Endpoint

1. Create `app/api/myapi.go`
2. Add handler:
   ```go
   package api
   func GetMyData(w http.ResponseWriter, r *http.Request) { ... }
   ```
3. Access at `GET /api/mydata`.

## Conventions

- Generated code goes to gonx/ (centralized, never edit manually)
- `_templ.go` files are legacy — being removed as we migrate to `.gonx`
- All template expressions are HTML-escaped by default
- Use `go:embed` or `public/` for static assets
