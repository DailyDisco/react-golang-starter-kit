package main

import (
	"log"
	"net/http"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	database.ConnectDB()

	// Create Chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware for React frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"}, // React dev server
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	setupRoutes(r)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setupRoutes(r chi.Router) {
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handlers.HealthCheck)

		// Example RESTful routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", handlers.GetUsers)          // GET /api/users
			r.Post("/", handlers.CreateUser)       // POST /api/users
			r.Get("/{id}", handlers.GetUser)       // GET /api/users/{id}
			r.Put("/{id}", handlers.UpdateUser)    // PUT /api/users/{id}
			r.Delete("/{id}", handlers.DeleteUser) // DELETE /api/users/{id}
		})
	})
}
