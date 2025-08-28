package main

import (
	"log"
	"net/http"
	"os"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/handlers"
	"react-golang-starter/internal/ratelimit"
	"strings"

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

	// Load rate limiting configuration
	rateLimitConfig := ratelimit.LoadConfig()
	if rateLimitConfig.Enabled {
		log.Println("Rate limiting is enabled")
	} else {
		log.Println("Rate limiting is disabled")
	}

	// Initialize database
	database.ConnectDB()

	// Create Chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Apply IP-based rate limiting globally
	r.Use(ratelimit.NewIPRateLimitMiddleware(rateLimitConfig))

	// CORS middleware for React frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   getAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	setupRoutes(r, rateLimitConfig)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setupRoutes(r chi.Router, rateLimitConfig *ratelimit.Config) {
	// Simple test route at root level
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Test route working!")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("API test route working!")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		})
		r.Get("/health", handlers.HealthCheck)

		// Authentication routes - combined public and protected
		r.Route("/auth", func(r chi.Router) {
			// Public authentication routes - stricter rate limiting
			r.Use(ratelimit.NewAuthRateLimitMiddleware(rateLimitConfig))
			r.Post("/register", auth.RegisterUser)           // POST /api/auth/register
			r.Post("/login", auth.LoginUser)                 // POST /api/auth/login
			r.Get("/verify-email", auth.VerifyEmail)         // GET /api/auth/verify-email
			r.Post("/reset-password", auth.RequestPasswordReset) // POST /api/auth/reset-password
			r.Post("/reset-password/confirm", auth.ResetPassword) // POST /api/auth/reset-password/confirm

			// Protected authentication routes - require authentication
			r.Route("/me", func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Get("/", auth.GetCurrentUser) // GET /api/auth/me
			})
		})

		// User routes - API rate limiting
		r.Route("/users", func(r chi.Router) {
			r.Use(ratelimit.NewAPIRateLimitMiddleware(rateLimitConfig))
			r.Get("/", handlers.GetUsers)          // GET /api/users (public for now)
			r.Post("/", handlers.CreateUser)       // POST /api/users (public for now)

			// Protected user routes
			r.Route("/{id}", func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Get("/", handlers.GetUser)       // GET /api/users/{id}
				r.Put("/", handlers.UpdateUser)    // PUT /api/users/{id}
				r.Delete("/", handlers.DeleteUser) // DELETE /api/users/{id}
			})
		})
	})

	// Swagger routes
	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/index.html")
	})
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})
}

// getAllowedOrigins returns the allowed CORS origins from environment variables
// Falls back to common development origins if not set
func getAllowedOrigins() []string {
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsEnv != "" {
		return strings.Split(originsEnv, ",")
	}

	// Default development origins
	return []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:8080",
		"*", // Allow all for development
	}
}
