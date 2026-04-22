# Framework Syntax & Usage Guide

This guide explains how to use the Go + HTMX Framework to build modern web applications.

## File-Based Routing

The framework uses the `app/pages` directory for routing:
- `app/pages/index.gonx` -> `/`
- `app/pages/about.gonx` -> `/about`
- `app/pages/user/_id.gonx` -> `/user/:id` (Dynamic Route)
- `app/pages/dashboard/index.gonx` -> `/dashboard`

## .gonx SFC (Single File Component)

A `.gonx` file consists of three optional blocks: `<script>`, `<template>`, and `<style>`.

### 1. Template Syntax
Use double curly braces for expressions and control flow:

```html
<template>
  <h1>Hello, {{ Name }}!</h1>
  
  {{ if IsAdmin }}
    <p>Welcome, Administrator.</p>
  {{ end }}
  
  <ul>
    {{ range Items }}
      <li>{{ _gonx_it }}</li>
    {{ end }}
  </ul>
</template>
```

### 2. Script Block (Go Logic)
The script block contains Go code. You can define variables and handlers.

#### Handlers (API Logic)
If you define functions starting with HTTP methods, they become routes:
- `func Get(c *fiber.Ctx)`: Custom GET handler for the page.
- `func Post(c *fiber.Ctx)`: Handles POST requests.

```html
<script>
  package dashboard
  
  import "playground/app/models"
  
  var Name = "User"
  
  func Post(c *fiber.Ctx) error {
      // Handle form submission
      return c.Redirect("/dashboard")
  }
</script>
```

### 3. Component Calls
Invoke components from `app/components/` using the `call` directive:

```html
{{ call components.Button(Type: "primary") }}
  Click Me
{{ end }}
```

## API Handlers

### Local API
Co-locate API logic in the same folder as the page:
- `app/pages/dashboard/api.go` -> Accessible at `/api/dashboard` (if it has `Get`, `Post`, etc.)

### Global API
Shared API handlers go in `app/api/`:
- `app/api/hello.go` -> `/api/hello`

## Layouts

Pages automatically use `app/layouts/index.gonx` by default. You can specify a different layout in the script block:

```go
// @layout layouts.Full("My Page Title")
func Index() {}
```

## Running the Project

1. **Build the framework**: `make build` (at root)
2. **Run in dev mode**: `make dev` (at root) or `../framework dev` (inside project)

The dev server features:
- **Hot-reload**: Automatic recompilation on file changes.
- **Auto-port**: Automatically picks next port if 3000 is busy.
- **Metrics**: Displays build time and bundle size for each page.
