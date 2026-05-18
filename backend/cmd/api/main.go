package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/pedropaffaro/deep-pothole-backend/internal/handlers"
)

func main() {
	godotenv.Load()
	app := fiber.New()

	// Middleware
	app.Use(cors.New())

	// Rotas
	api := app.Group("/api/v1")
	api.Post("/complaint", handlers.CreateComplaint)

	log.Fatal(app.Listen(":8080"))
}