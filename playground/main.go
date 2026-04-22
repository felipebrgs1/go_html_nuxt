package main

import (
	"fmt"
	"os"

	framework "playground/gonx/framework_gen"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

func main() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: os.Getenv("GONX_RESTART") == "true",
	})

	// Brotli/Gzip compression only in production
	if os.Getenv("GO_ENV") == "production" {
		app.Use(compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}))
	}

	// Static files with long-term cache
	app.Static("/", "./public", fiber.Static{
		MaxAge:   31536000, // 1 year
		Compress: os.Getenv("GO_ENV") == "production",
	})

	// Register generated routes
	framework.RegisterRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server running on http://localhost:%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
