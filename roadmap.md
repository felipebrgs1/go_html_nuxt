# Roadmap: Framework Web "Nuxt-like" em Go + HTMX

**Visão:** Construir um framework web de alta performance que una a velocidade e tipagem do Go com a produtividade do Nuxt.js (file-based routing, zero-config build) e a reatividade do Phoenix LiveView (usando HTMX e Templ para type-safe HTML).

## Fase 1: Fundação do CLI e Hot-Reload ✅ CONCLUÍDA
O coração do framework não é apenas o código, mas a **ferramenta de linha de comando (CLI)** que proporciona a "Developer Experience" mágica.

- [x] **Setup Inicial:** Criar a estrutura do repositório (separando o código do CLI e o código da biblioteca runtime).
- [x] **Comando `dev`:** Criar a fundação do comando `framework dev`.
- [x] **Hot-Reloading (Watcher):** Implementar um file watcher que detecta mudanças em arquivos `.go` e `.templ` e reinicia o servidor instantaneamente (substituindo a necessidade do `air`).

**Como usar:**
```bash
go build ./cmd/framework
./framework dev
```

## Fase 2: Roteamento Baseado em Arquivos (File-based Routing) ✅ CONCLUÍDA
Fazer o Go entender a estrutura de pastas e gerar as rotas dinamicamente, igual ao Nuxt.

- [x] **Definição de Padrões:** Estabelecer a convenção de pastas (ex: `app/pages/index.go`, `app/pages/users/_id/index.go`).
- [x] **Análise de AST:** Usar o pacote `go/ast` para ler as funções dentro de `app/pages` em tempo de desenvolvimento.
- [x] **Code Generation:** O CLI gera automaticamente `.framework/routes.gen.go` que mapeia arquivos físicos para o roteador do Go (`net/http` do Go 1.22+).
- [x] **Suporte a API Routes:** Permitir a criação de rotas JSON nativas na pasta `app/api/`.

**Convenções implementadas:**
- `app/pages/index.go` → `GET /`
- `app/pages/users/_id/index.go` → `GET /users/{id}` (`_param` = parâmetro dinâmico)
- `app/api/hello.go` → `/api/hello` (método detectado pelo nome: `GetHello` → `GET`)

**Como usar:**
```bash
./framework dev
# O arquivo .framework/routes.gen.go é gerado automaticamente
# Importe-o no seu main.go:
#   "seu-modulo/.framework"
#   framework.RegisterRoutes(mux)
```

## Fase 3: Integração da View Engine (Templ + HTMX) ✅ CONCLUÍDA
Garantir que a criação de interfaces seja tipada, rápida e reativa.

- [x] **Integração com Templ:** O CLI compila automaticamente os arquivos `.templ` com `templ generate` antes de iniciar/reiniciar o servidor.
- [x] **HTMX Built-in:** Script HTMX embutido no layout base + helpers Go (`htmx.IsHTMXRequest`, `htmx.Redirect`, `htmx.PushURL`, etc).
- [x] **Sistema de Layouts:** Layout base em `app/layouts/base.templ` com slots para injeção de conteúdo via `children...`.

**Stack visual funcionando:**
- `app/layouts/base.templ` → Layout HTML com HTMX + Tailwind
- `app/pages/index.templ` → View tipada injetada no layout
- `app/pages/index.go` → Handler Go que renderiza o template

## Fase 4: DX (Developer Experience) Avançada
Funcionalidades que deixam os desenvolvedores felizes e produtivos.

- [x] **Gerenciamento de Assets:** Arquivos estáticos da pasta `public/` servidos automaticamente (com fallback para rotas geradas).
- [x] **Integração com TailwindCSS:** Suporte embutido via CLI standalone. O comando `dev` compila `public/input.css` → `public/styles.css` automaticamente.
- [ ] **Middlewares Mágicos:** Capacidade de adicionar middlewares locais baseados em pastas (ex: `app/pages/admin/middleware.go` protege toda a rota `/admin`).
- [ ] **State & Context Helpers:** Facilitadores para lidar com sessão, cookies e injeção de dependências nas rotas.

## Fase 5: Preparação para Produção (Build System)
Garantir que o resultado final seja o padrão ouro do Go: um único binário minúsculo.

- [ ] **Comando `build`:** Compila os templates, minifica assets (Tailwind), e executa o `go build` otimizado gerando apenas um arquivo executável.
- [ ] **Embed FS:** Usar o `go:embed` para empacotar o HTML base e assets finais dentro do próprio binário, facilitando o deploy via Docker ou bare-metal.

---

### Estrutura de Pastas Alvo (Como o projeto do usuário final vai parecer):
```text
meu-projeto/
├── app/
│   ├── layouts/
│   │   └── base.templ        # Layout base (HTML, HEAD, BODY)
│   ├── pages/
│   │   ├── index.go          # Lógica da página inicial (Go)
│   │   ├── index.templ       # View da página inicial (Templ + HTMX)
│   │   └── users/
│   │       └── _id/
│   │           ├── index.go    # Rota dinâmica /users/123
│   │           └── index.templ
│   └── api/
│       └── webhooks.go       # Rotas de API puro (JSON)
├── public/
│   └── favicon.ico           # Assets estáticos
├── go.mod
└── tailwind.config.js        # Configuração do Tailwind (opcional)
```
