package main

import (
	"log"
	"net/http"
	"os"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/handlers"
	"react-golang-starter/internal/middleware"
	"react-golang-starter/internal/ratelimit"
	"react-golang-starter/internal/services"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zerologlog "github.com/rs/zerolog/log"
)

// @title React Go Starter Kit API
// @version 1.0.0
// @description A comprehensive REST API for the React Go Starter Kit application built with Fiber, GORM, and PostgreSQL. This API provides secure user authentication, user management, and system health monitoring.
//
// ## Features
//
// - **User Authentication**: JWT-based authentication with email verification
// - **User Management**: Complete CRUD operations for user accounts
// - **Password Security**: Secure password hashing and reset functionality
// - **Rate Limiting**: Built-in protection against abuse
// - **Health Monitoring**: System health checks and status endpoints
//
// ## Authentication
//
// Most endpoints require JWT Bearer token authentication. Obtain a token by logging in
// and include it in the Authorization header: `Authorization: Bearer {token}`
//
// ## Rate Limiting
//
// API endpoints are protected by rate limiting to prevent abuse. Different endpoints
// have different rate limits based on their sensitivity.
//
// @termsOfService https://github.com/your-org/react-golang-starter-kit
//
// @contact.name API Support
// @contact.url https://github.com/your-org/react-golang-starter-kit/issues
// @contact.email support@example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Format: `Authorization: Bearer {token}`
// To obtain a token: 1. Register via POST /api/auth/register 2. Login via POST /api/auth/login
//
// @tag.name auth
// @tag.description User authentication and authorization endpoints including login, registration, password reset, and email verification
//
// @tag.name users
// @tag.description User management operations including CRUD operations for user accounts
//
// @tag.name health
// @tag.description System health monitoring and status endpoints for checking server availability
func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		zerologlog.Info().Msg("No .env file found, using system environment variables")
	}

	// Load logging configuration
	logConfig := middleware.LoadLogConfig()

	// Configure zerolog for structured logging
	zerolog.TimeFieldFormat = logConfig.TimeFormat

	// Set log level based on configuration
	switch strings.ToLower(logConfig.Level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Configure pretty printing if enabled
	if logConfig.Pretty {
		zerologlog.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	}

	// Log configuration status
	if logConfig.Enabled {
		zerologlog.Info().
			Str("level", logConfig.Level).
			Bool("user_context", logConfig.IncludeUserContext).
			Bool("request_body", logConfig.IncludeRequestBody).
			Bool("response_body", logConfig.IncludeResponseBody).
			Float64("sampling_rate", logConfig.SamplingRate).
			Msg("structured logging enabled")
	} else {
		zerologlog.Info().Msg("structured logging disabled")
	}

	// Load rate limiting configuration
	rateLimitConfig := ratelimit.LoadConfig()
	if rateLimitConfig.Enabled {
		zerologlog.Info().Msg("rate limiting enabled")
	} else {
		zerologlog.Info().Msg("rate limiting disabled")
	}

	// Initialize database
	database.ConnectDB()

	// Initialize Redis cache
	redisClient := cache.ConnectRedis()
	var cacheService *cache.Service
	if redisClient != nil {
		cacheService = cache.NewService(cache.NewCache(redisClient))
		defer redisClient.Close()
	} else {
		zerologlog.Warn().Msg("Redis not available, cache service disabled")
		cacheService = nil
	}

	// Create Chi router
	r := chi.NewRouter()

	// Global middleware
	// Compression middleware for improved performance (must be first)
	r.Use(chimiddleware.Compress(5, "application/json", "text/plain", "text/html"))

	r.Use(middleware.StructuredLoggerWithConfig(logConfig))
	r.Use(chimiddleware.Recoverer)

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

	// Initialize file service
	fileService, err := services.NewFileService()
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("failed to initialize file service")
	}

	// Initialize the service with dependencies
	appService := handlers.NewService(redisClient)

	// Health check at root level for Docker health checks
	r.Get("/health", appService.HealthCheck)

	// Routes
	setupRoutes(r, rateLimitConfig, cacheService, appService, fileService)

	zerologlog.Info().Str("port", ":8080").Msg("server starting")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setupRoutes(r chi.Router, rateLimitConfig *ratelimit.Config, cacheService *cache.Service, appService *handlers.Service, fileService *services.FileService) {
	// Simple test route at root level
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte("Test route working!")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("API test route working!")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		})
		r.Get("/health", appService.HealthCheck)

		// File upload routes
		r.Route("/files", func(r chi.Router) {
			r.Use(ratelimit.NewAPIRateLimitMiddleware(rateLimitConfig))

			// File upload - requires authentication for security
			r.Route("/upload", func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Post("/", handlers.NewFileHandler(fileService).UploadFile) // POST /api/files/upload
			})

			// File operations - public access for downloads, authenticated for management
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/download", handlers.NewFileHandler(fileService).DownloadFile) // GET /api/files/{id}/download
				r.Get("/url", handlers.NewFileHandler(fileService).GetFileURL)        // GET /api/files/{id}/url
				r.Get("/", handlers.NewFileHandler(fileService).GetFileInfo)          // GET /api/files/{id}

				// Protected operations - require authentication
				r.Route("/", func(r chi.Router) {
					r.Use(auth.AuthMiddleware)
					r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
					r.Delete("/", handlers.NewFileHandler(fileService).DeleteFile) // DELETE /api/files/{id}
				})
			})

			// List files - requires authentication
			r.Route("/", func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Get("/", handlers.NewFileHandler(fileService).ListFiles) // GET /api/files
			})

			// Storage status - public endpoint
			r.Get("/storage/status", handlers.NewFileHandler(fileService).GetStorageStatus) // GET /api/files/storage/status
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

	// Default development origins - no wildcard to allow credentials
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3002",
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
}
