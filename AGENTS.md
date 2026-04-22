# Go Framework — Internal Guide

## Core Mission
Build a high-performance, Nuxt-inspired framework using **Go + HTMX + TailwindCSS**. The focus is on the **compiler** (`.gonx` -> `.go`) and a **decoupled CLI** that treats projects as black boxes.

For usage guidelines and template syntax, see [SYNTAX.md](./SYNTAX.md).

## Decoupled Architecture
The framework is designed to be independent of the application:
- **Framework (Root)**: Contains the core engine (`pkg/`) and the CLI wrapper (`scripts/`).
- **Application (e.g., playground/)**: A separate Go module that uses the framework to generate UI and routes.
- **Generic Scanner**: The router scanner automatically detects the module name and paths relative to the project root.

## Framework Components (pkg/)
- **gonx/**: The SFC (Single File Component) engine.
  - `parser.go`: Extracts `<template>`, `<script>`, and `<style>` blocks.
  - `compiler.go`: Transforms template syntax into standard Go `io.Writer` calls.
- **router/**:
  - `scanner.go`: Recursively scans `app/pages` and `app/api`. Maps file paths to routes.
  - **Precedence**: Custom handlers (`Get`, `Post`, etc.) in `.gonx` scripts take precedence over default rendering.
- **generator/**: Produces `gonx/framework_gen/routes.gen.go`.
- **cli/**: Implements `dev` and `lint` commands with multi-project support and auto-port discovery.

## The `.gonx` Compiler Logic
1. **Extraction**: Regex-based block extraction.
2. **AST Analysis**: The `<script>` block is parsed via `go/parser`.
   - **Auto-Imports**: Automatically adds correct module imports based on internal package usage.
3. **Template Compilation**:
   - `{{ expr }}` -> HTML-escaped string interpolation.
   - `{{ call Component }}` -> Helper for nested component rendering with slots.

## Project Structure (Generic App)
- `app/`: User-space code.
  - `pages/`: UI pages. Can contain local `api.go` files.
  - `api/`: General API handlers.
  - `components/`: Reusable UI components.
- `gonx/`: **Generated** code (managed by framework).
- `main.go`: Application entry point (registers generated routes).

## Key Implementation Rules
- **Decoupling**: The framework must NOT have hardcoded references to project folder names (like "playground").
- **Efficiency**: Only show the first 10 compiled files in logs; track build times and bundle sizes.
- **Safety**: Auto-port discovery (3000 -> 3001...) ensures multiple instances can run without collision.
