package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go_template/pkg/htmx"
)

func GetHello(c *fiber.Ctx) error {
	// Se for HTMX, retorna HTML formatado
	if htmx.IsFiberHTMXRequest(c) {
		c.Type("html")
		return c.SendString(fmt.Sprintf(`<div class="p-3 bg-green-50 border border-green-200 rounded-lg">
			<div class="flex items-center gap-2">
				<div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
				<span class="text-green-800 font-semibold text-sm">Sucesso!</span>
			</div>
			<p class="text-green-700 text-sm mt-1">%s</p>
			<p class="text-green-600 text-xs mt-1">Resposta HTMX em %s</p>
		</div>`,
			"Hello from API via HTMX! 🚀",
			c.Get("HX-Request")))
	}

	// API pura retorna JSON
	return c.JSON(map[string]string{
		"message": "Hello from API",
	})
}
