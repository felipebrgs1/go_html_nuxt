package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go_template/pkg/htmx"
)

func GetHello(w http.ResponseWriter, r *http.Request) {
	// Se for HTMX, retorna HTML formatado
	if htmx.IsHTMXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="p-3 bg-green-50 border border-green-200 rounded-lg">
			<div class="flex items-center gap-2">
				<div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
				<span class="text-green-800 font-semibold text-sm">Sucesso!</span>
			</div>
			<p class="text-green-700 text-sm mt-1">%s</p>
			<p class="text-green-600 text-xs mt-1">Resposta HTMX em %s</p>
		</div>`,
			"Hello from API via HTMX! 🚀",
			r.Header.Get("HX-Request"))
		return
	}

	// API pura retorna JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from API",
	})
}
