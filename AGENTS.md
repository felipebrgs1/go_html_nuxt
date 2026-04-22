# Go Framework — Internal Guide

## Core Mission
Build a high-performance, Nuxt-inspired framework using **Go + HTMX + TailwindCSS**. The focus is on the **compiler** (`.gonx` -> `.go`) and the **build orchestration**.

## Framework Components (pkg/)

- **gonx/**: The SFC (Single File Component) engine.
  - `parser.go`: Extracts `<template>`, `<script>`, and `<style>` (optional) blocks. Uses `go/ast` to analyze the script.
  - `compiler.go`: Transforms template syntax (range, if, interpolation) into standard Go `io.Writer` calls.
  - `generator.go`: Manages the output of compiled files into the `gonx/` directory.
- **router/**:
  - `scanner.go`: Recursively scans `app/pages` and `app/api`. Maps file paths to HTTP routes (e.g., `_id.gonx` -> `:id`).
- **generator/**:
  - `generator.go`: Produces `gonx/framework_gen/routes.gen.go`. It centralizes route registration using Fiber.
- **watcher/**:
  - `watcher.go`: Monitors `.go`, `.gonx`, `.templ`, and `.css` files. Triggers the rebuild pipeline and restarts the dev server.

## The `.gonx` Compiler Logic

1. **Extraction**: Regex-based block extraction.
2. **AST Analysis**: The `<script>` block is parsed via `go/parser`. 
   - Identifies the package name.
   - Discovers exported handler functions.
   - **Auto-Imports**: Scans code for internal package usage (e.g., `layouts.Base`) and automatically adds correct module imports.
3. **Template Compilation**:
   - `{{ expr }}` -> `html.EscapeString(fmt.Sprintf("%v", expr))`
   - `{{ if/range/end }}` -> Native Go control flow.
   - `{{ call Component }}` -> Helper for nested component rendering.

## Build Pipeline (Dev Mode)

When a file changes, the CLI (`cmd/framework`) executes:
1. **Gonx Compile**: All `.gonx` files -> `gonx/` Go files.
2. **Templ Generate**: (Legacy/Current) Compile `.templ` files.
3. **Tailwind Build**: Compile CSS to `public/styles.css`.
4. **Route Generation**: Update `gonx/framework_gen/routes.gen.go` based on the latest scan.
5. **Hot Restart**: Re-run `go run main.go`.

## Project Structure
- `cmd/framework/`: The CLI entry point.
- `app/`: User-space code.
  - `pages/`: UI pages. Can contain local `api.go` files (e.g., `dashboard/api.go` -> `/api/dashboard`).
  - `api/`: General API handlers (.go).
  - `layouts/`: Shared layouts.
- `gonx/`: **Generated** code.
  - `app/`: Compiled SFC components.
  - `framework_gen/`: Generated route registry.
- `pkg/`: **Framework Core** (the actual logic we are developing).

## Key Implementation Rules
- **No Dots**: Generated packages must NOT start with `.` (e.g., use `gonx/` instead of `.gonx/`) to ensure Go module compatibility.
- **Efficiency**: Avoid full project rebuilds if only assets changed.
- **Type Safety**: The compiler must ensure that variables used in templates are valid within the script block's scope.
