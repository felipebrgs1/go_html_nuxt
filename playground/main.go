package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"playground/gonx/framework_gen"
)

func main() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	// Static files
	app.Static("/", "./public")

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
