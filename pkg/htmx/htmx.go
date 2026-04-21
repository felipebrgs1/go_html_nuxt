package htmx

import (
	"net/http"
	"github.com/gofiber/fiber/v2"
)

// IsHTMXRequest retorna true se a requisição veio do HTMX (net/http)
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// IsFiberHTMXRequest retorna true se a requisição veio do HTMX (Fiber)
func IsFiberHTMXRequest(c *fiber.Ctx) bool {
	return c.Get("HX-Request") == "true"
}

// IsHTMXBoosted retorna true se a requisição foi feita via hx-boost
func IsHTMXBoosted(r *http.Request) bool {
	return r.Header.Get("HX-Boosted") == "true"
}

// FiberRedirect envia um header HX-Redirect para navegação full-page via HTMX
func FiberRedirect(c *fiber.Ctx, url string) {
	c.Set("HX-Redirect", url)
}

// Redirect envia um header HX-Redirect para navegação full-page via HTMX (net/http)
func Redirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
}

// PushURL envia um header HX-Push-Url para atualizar o browser URL
func PushURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Push-Url", url)
}

// FiberTrigger dispara um evento HTMX no cliente (Fiber)
func FiberTrigger(c *fiber.Ctx, event string) {
	c.Set("HX-Trigger", event)
}

// Trigger dispara um evento HTMX no cliente
func Trigger(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger", event)
}
