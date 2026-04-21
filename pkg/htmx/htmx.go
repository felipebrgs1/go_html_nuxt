package htmx

import "net/http"

// IsHTMXRequest retorna true se a requisição veio do HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// IsHTMXBoosted retorna true se a requisição foi feita via hx-boost
func IsHTMXBoosted(r *http.Request) bool {
	return r.Header.Get("HX-Boosted") == "true"
}

// Redirect envia um header HX-Redirect para navegação full-page via HTMX
func Redirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
}

// PushURL envia um header HX-Push-Url para atualizar o browser URL
func PushURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Push-Url", url)
}

// Retarget altera o alvo do swap HTMX
func Retarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

// Reswap altera o método de swap
func Reswap(w http.ResponseWriter, strategy string) {
	w.Header().Set("HX-Reswap", strategy)
}

// Trigger dispara um evento HTMX no cliente
func Trigger(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger", event)
}
