package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"react-golang-starter/internal/ai"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/config"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/email"
	"react-golang-starter/internal/handlers"
	"react-golang-starter/internal/jobs"
	"react-golang-starter/internal/middleware"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/ratelimit"
	"react-golang-starter/internal/services"
	"react-golang-starter/internal/stripe"
	"react-golang-starter/internal/websocket"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
// @BasePath /api/v1
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
// validateProductionSecrets validates that critical secrets are properly configured in production.
// This prevents deployment with weak or default secrets that could compromise security.
func validateProductionSecrets() {
	env := os.Getenv("GO_ENV")
	if env != "production" && env != "prod" {
		return // Only validate in production
	}

	// Validate JWT_SECRET
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		zerologlog.Fatal().Msg("JWT_SECRET is required in production")
	}
	if len(jwtSecret) < 32 {
		zerologlog.Fatal().Msg("JWT_SECRET must be at least 32 characters in production")
	}
	// Check for known default/weak secrets
	weakSecrets := []string{
		"dev-jwt-secret-key-change-in-production",
		"secret",
		"jwt-secret",
		"changeme",
		"your-secret-key",
	}
	for _, weak := range weakSecrets {
		if strings.EqualFold(jwtSecret, weak) {
			zerologlog.Fatal().Msg("Using a known weak JWT_SECRET in production is forbidden")
		}
	}

	// Validate Redis password if Redis is configured
	redisURL := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisURL != "" && redisPassword == "" {
		zerologlog.Warn().Msg("REDIS_PASSWORD is empty - Redis should be password-protected in production")
	}

	// Validate database password
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		zerologlog.Fatal().Msg("DB_PASSWORD is required in production")
	}
	weakDBPasswords := []string{"devpass", "password", "postgres", "admin", "123456"}
	for _, weak := range weakDBPasswords {
		if strings.EqualFold(dbPassword, weak) {
			zerologlog.Fatal().Msg("Using a known weak DB_PASSWORD in production is forbidden")
		}
	}

	// Validate debug mode is disabled
	if os.Getenv("DEBUG") == "true" {
		zerologlog.Fatal().Msg("DEBUG mode must be disabled in production (DEBUG=false)")
	}

	zerologlog.Info().Msg("production secrets validation passed")
}

// HubBroadcaster adapts websocket.Hub to implement cache.CacheBroadcaster interface.
// This breaks the import cycle between cache and websocket packages.
type HubBroadcaster struct {
	hub *websocket.Hub
}

// BroadcastCacheInvalidation implements cache.CacheBroadcaster interface
func (h *HubBroadcaster) BroadcastCacheInvalidation(payload cache.CacheInvalidatePayload) {
	// Convert to websocket payload format
	wsPayload := websocket.CacheInvalidatePayload{
		QueryKeys: payload.QueryKeys,
		Event:     payload.Event,
		Timestamp: payload.Timestamp,
	}
	h.hub.Broadcast(websocket.MessageTypeCacheInvalidate, wsPayload)
}

func main() {
	// Load environment variables - try multiple locations
	err := godotenv.Load(".env.local")
	if err != nil {
		// Try .env as fallback (for when running from backend/ directory)
		err = godotenv.Load(".env")
		if err != nil {
			zerologlog.Info().Msg("No .env.local or .env file found, using system environment variables")
		}
	}

	// Validate production secrets before proceeding
	validateProductionSecrets()

	// Initialize Sentry for error tracking (optional - controlled by SENTRY_DSN env var)
	sentryConfig := middleware.LoadSentryConfig()
	if err := middleware.InitSentry(sentryConfig); err != nil {
		zerologlog.Warn().Err(err).Msg("Sentry initialization failed, continuing without error tracking")
	}
	// Ensure Sentry flushes before exit
	defer middleware.FlushSentry(2 * time.Second)

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

	// Seed initial feature flags (idempotent - only creates if not exist)
	if err := database.SeedFeatureFlags(); err != nil {
		zerologlog.Warn().Err(err).Msg("Feature flag seeding failed")
	}

	// Auto-seed database in development when AUTO_SEED=true
	seedConfig := database.LoadSeedConfig()
	if seedConfig.Enabled {
		zerologlog.Info().Msg("AUTO_SEED enabled, seeding database...")
		if err := database.SeedAll(seedConfig); err != nil {
			zerologlog.Warn().Err(err).Msg("Auto-seeding failed")
		}
	}

	// Initialize cache
	cacheConfig := cache.LoadConfig()
	if err := cache.Initialize(cacheConfig); err != nil {
		zerologlog.Warn().Err(err).Msg("cache initialization failed, continuing without cache")
	} else if cacheConfig.Enabled {
		zerologlog.Info().
			Str("type", cacheConfig.Type).
			Bool("available", cache.IsAvailable()).
			Msg("cache initialized")
	}
	defer cache.Close()

	// Initialize email service
	emailConfig := email.LoadConfig()
	if err := email.Initialize(emailConfig); err != nil {
		zerologlog.Warn().Err(err).Msg("email initialization failed, continuing without email")
	} else if emailConfig.Enabled {
		zerologlog.Info().
			Str("host", emailConfig.SMTPHost).
			Bool("dev_mode", emailConfig.DevMode).
			Msg("email service initialized")
	}
	defer email.Close()

	// Initialize Stripe service
	stripeConfig := stripe.LoadConfig()
	if err := stripe.Initialize(stripeConfig); err != nil {
		zerologlog.Warn().Err(err).Msg("stripe initialization failed, continuing without billing")
	} else if stripeConfig.Enabled {
		zerologlog.Info().
			Bool("available", stripe.IsAvailable()).
			Msg("stripe service initialized")
	}

	// Initialize AI service (Gemini)
	aiConfig := ai.LoadConfig()
	if err := ai.Initialize(aiConfig); err != nil {
		zerologlog.Warn().Err(err).Msg("ai initialization failed, continuing without AI features")
	} else if aiConfig.Enabled {
		zerologlog.Info().
			Str("model", aiConfig.Model).
			Bool("available", ai.IsAvailable()).
			Msg("ai service initialized")
	}

	// Initialize job queue (River)
	jobsConfig := jobs.LoadConfig()
	if err := jobs.Initialize(jobsConfig); err != nil {
		zerologlog.Warn().Err(err).Msg("job system initialization failed, continuing without jobs")
	} else if jobsConfig.Enabled {
		zerologlog.Info().
			Int("workers", jobsConfig.WorkerCount).
			Bool("available", jobs.IsAvailable()).
			Msg("job system initialized")
	}

	// Initialize metrics retention cleanup (runs on startup and every 24 hours)
	retentionConfig := jobs.LoadMetricsRetentionConfig()
	if retentionConfig.Enabled {
		zerologlog.Info().
			Int("retention_days", retentionConfig.RetentionDays).
			Msg("metrics retention cleanup enabled")
	}

	// Initialize WebSocket hub for real-time communication
	wsHub := websocket.NewHub()
	zerologlog.Info().Msg("WebSocket hub initialized")

	// Initialize cache broadcaster for real-time cache invalidation via WebSocket
	cache.InitBroadcaster(&HubBroadcaster{hub: wsHub})

	// Create Chi router
	r := chi.NewRouter()

	// Global middleware
	// Compression middleware for improved performance (must be first)
	r.Use(chimiddleware.Compress(5, "application/json", "text/plain", "text/html"))

	// Prometheus metrics middleware (early in chain to capture all requests)
	r.Use(middleware.PrometheusMiddleware)

	// Request ID middleware (before logger so IDs are logged)
	r.Use(middleware.RequestIDMiddleware)

	// Sentry middleware (captures panics and reports errors)
	r.Use(middleware.SentryMiddleware(sentryConfig))

	r.Use(middleware.StructuredLoggerWithConfig(logConfig))
	r.Use(middleware.RecoveryMiddleware)

	// CORS middleware for React frontend (MUST be before rate limiting so preflight OPTIONS requests work)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.GetAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With", "X-Request-ID", "Origin"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Apply IP-based rate limiting globally (after CORS to allow preflight requests)
	r.Use(ratelimit.NewIPRateLimitMiddleware(rateLimitConfig))

	// Security headers middleware
	securityConfig := middleware.LoadSecurityConfig()
	r.Use(middleware.SecurityHeaders(securityConfig))
	if securityConfig.Enabled {
		zerologlog.Info().Msg("security headers enabled")
	}

	// Cache headers middleware (applies Cache-Control based on route patterns)
	cacheHeadersConfig := middleware.LoadCacheHeadersConfig()
	r.Use(middleware.CacheHeaders(cacheHeadersConfig))
	if cacheHeadersConfig.Enabled {
		zerologlog.Info().Msg("cache headers middleware enabled")
	}

	// CSRF protection middleware
	csrfConfig := middleware.LoadCSRFConfig()
	r.Use(middleware.CSRFProtection(csrfConfig))
	if csrfConfig.Enabled {
		zerologlog.Info().Msg("CSRF protection enabled")
	}

	// Request body size limit middleware (prevents memory exhaustion attacks)
	r.Use(middleware.MaxBodySize(middleware.DefaultMaxBodySize))
	zerologlog.Info().Int64("max_bytes", middleware.DefaultMaxBodySize).Msg("request body size limit enabled")

	// Initialize OAuth
	auth.InitOAuth()
	if auth.IsOAuthConfigured() {
		zerologlog.Info().
			Bool("google", auth.IsGoogleOAuthConfigured()).
			Bool("github", auth.IsGitHubOAuthConfigured()).
			Msg("OAuth providers initialized")
	}

	// Initialize file service
	fileService, err := services.NewFileService()
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("failed to initialize file service")
	}

	// Initialize the service with dependencies
	appService := handlers.NewService()

	// Health check at root level for Docker health checks
	r.Get("/health", appService.HealthCheck)

	// Prometheus metrics endpoint (internal, no auth required)
	r.Handle("/metrics", promhttp.Handler())

	// WebSocket endpoint (before other routes, no rate limiting)
	r.Get("/ws", websocket.Handler(wsHub))

	// Routes
	setupRoutes(r, rateLimitConfig, stripeConfig, appService, fileService, wsHub)

	// Create server with timeouts to prevent slowloris and other DoS attacks
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Higher for file uploads and streaming
		IdleTimeout:  120 * time.Second,
	}

	// Start job processing in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start WebSocket hub
	go wsHub.Run(ctx)

	if jobs.IsAvailable() {
		if err := jobs.Start(ctx); err != nil {
			zerologlog.Fatal().Err(err).Msg("failed to start job processing")
		}
	}

	// Start periodic metrics retention cleanup
	jobs.StartPeriodicRetention(ctx, retentionConfig)

	// Graceful shutdown handling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		zerologlog.Info().Msg("shutdown signal received, starting graceful shutdown")

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Stop WebSocket hub first
		wsHub.Stop()
		zerologlog.Info().Msg("WebSocket hub stopped")

		// Stop job processing
		if jobs.IsAvailable() {
			if err := jobs.Stop(shutdownCtx); err != nil {
				zerologlog.Error().Err(err).Msg("error stopping job system")
			}
		}

		// Shutdown HTTP server
		if err := server.Shutdown(shutdownCtx); err != nil {
			zerologlog.Error().Err(err).Msg("error shutting down server")
		}

		cancel()
	}()

	zerologlog.Info().Str("port", ":8080").Msg("server starting")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server error:", err)
	}

	zerologlog.Info().Msg("server stopped gracefully")
}

func setupRoutes(r chi.Router, rateLimitConfig *ratelimit.Config, stripeConfig *stripe.Config, appService *handlers.Service, fileService *services.FileService, wsHub *websocket.Hub) {
	// Simple test route at root level
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte("Test route working!")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
	})

	// API routes setup - shared between /api and /api/v1
	apiRoutes := func(r chi.Router) {
		// setupAPIRoutes must be called FIRST because it registers middleware with r.Use()
		// Chi requires all middleware to be defined before any routes
		setupAPIRoutes(r, rateLimitConfig, stripeConfig, appService, fileService, wsHub)

		// These routes come after setupAPIRoutes to ensure middleware is registered first
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("API test route working!")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		})
		r.Get("/health", appService.HealthCheck)
	}

	// Mount API routes
	// /api/v1 is the canonical versioned endpoint
	// /api is kept for backwards compatibility (points to same handlers)
	r.Route("/api/v1", apiRoutes)
	r.Route("/api", apiRoutes)

	// Swagger routes
	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/index.html")
	})
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})
	// Serve swagger static files (self-hosted swagger-ui)
	staticFs := http.FileServer(http.Dir("docs/static"))
	r.Handle("/swagger/static/*", http.StripPrefix("/swagger/static/", staticFs))
}

// setupAPIRoutes configures all API endpoints
func setupAPIRoutes(r chi.Router, rateLimitConfig *ratelimit.Config, stripeConfig *stripe.Config, appService *handlers.Service, fileService *services.FileService, wsHub *websocket.Hub) {
	// Initialize organization service and handlers
	orgService := services.NewOrgService(database.DB)
	orgHandler := handlers.NewOrgHandler(orgService)
	tenantMiddleware := auth.NewTenantMiddleware(database.DB)

	// Initialize usage service and handlers
	usageService := services.NewUsageService(database.DB)
	usageService.SetHub(wsHub) // Enable WebSocket alerts
	usageHandler := handlers.NewUsageHandler(usageService)

	// Usage metering middleware (records API calls for authenticated users)
	r.Use(middleware.UsageMiddleware(usageService))

	// CSRF token endpoint - allows frontend to get a fresh CSRF token
	csrfConfig := middleware.LoadCSRFConfig()
	r.Get("/csrf-token", middleware.GetCSRFToken(csrfConfig))

	// Authentication routes
	r.Route("/auth", func(r chi.Router) {
		// Strict rate limiting for login/register/password reset (brute-force protection)
		r.Group(func(r chi.Router) {
			r.Use(ratelimit.NewAuthRateLimitMiddleware(rateLimitConfig))
			r.Post("/register", auth.RegisterUser)                // POST /api/auth/register
			r.Post("/login", auth.LoginUser)                      // POST /api/auth/login
			r.Post("/reset-password", auth.RequestPasswordReset)  // POST /api/auth/reset-password
			r.Post("/reset-password/confirm", auth.ResetPassword) // POST /api/auth/reset-password/confirm
			r.Get("/verify-email", auth.VerifyEmail)              // GET /api/auth/verify-email
		})

		// Token refresh uses more lenient API rate limit (called automatically by frontend)
		r.Group(func(r chi.Router) {
			r.Use(ratelimit.NewAPIRateLimitMiddleware(rateLimitConfig))
			r.Post("/refresh", auth.RefreshAccessToken) // POST /api/auth/refresh - Exchange refresh token for new access token
			r.Post("/logout", auth.LogoutUser)          // POST /api/auth/logout
		})

		// Protected auth endpoints
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
			r.Get("/me", auth.GetCurrentUser) // GET /api/auth/me
		})

		// OAuth routes
		r.Route("/oauth", func(r chi.Router) {
			// Public OAuth endpoints (use auth rate limit)
			r.Group(func(r chi.Router) {
				r.Use(ratelimit.NewAuthRateLimitMiddleware(rateLimitConfig))
				r.Get("/{provider}", auth.GetOAuthURL)                  // GET /api/auth/oauth/{provider} - Get OAuth URL
				r.Get("/{provider}/callback", auth.HandleOAuthCallback) // GET /api/auth/oauth/{provider}/callback - OAuth callback
			})

			// Protected OAuth endpoints
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Get("/providers", auth.GetLinkedProviders) // GET /api/auth/oauth/providers - List linked providers
				r.Delete("/{provider}", auth.UnlinkProvider) // DELETE /api/auth/oauth/{provider} - Unlink provider
			})
		})
	})

	// User management routes
	r.Route("/users", func(r chi.Router) {
		r.Use(ratelimit.NewAPIRateLimitMiddleware(rateLimitConfig))

		// Public routes
		r.Get("/", handlers.GetUsers()) // GET /api/users - List all users

		// Admin-only routes - require admin privileges
		r.Group(func(r chi.Router) {
			r.Use(auth.AdminMiddleware)        // Requires admin or super_admin role
			r.Post("/", handlers.CreateUser()) // POST /api/users - Create new user (admin only)
		})

		// Protected routes - require authentication (must be before /{id} to avoid "me" being matched as ID)
		r.Route("/me", func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))

			r.Get("/", handlers.GetCurrentUser())    // GET /api/users/me - Get current user
			r.Put("/", handlers.UpdateCurrentUser()) // PUT /api/users/me - Update current user

			// User preferences
			r.Get("/preferences", handlers.GetUserPreferences)    // GET /api/users/me/preferences
			r.Put("/preferences", handlers.UpdateUserPreferences) // PUT /api/users/me/preferences

			// Session management
			r.Get("/sessions", handlers.GetUserSessions)       // GET /api/users/me/sessions
			r.Delete("/sessions", handlers.RevokeAllSessions)  // DELETE /api/users/me/sessions - Revoke all other sessions
			r.Delete("/sessions/{id}", handlers.RevokeSession) // DELETE /api/users/me/sessions/{id}

			// Login history
			r.Get("/login-history", handlers.GetLoginHistory) // GET /api/users/me/login-history

			// Password change
			r.Put("/password", handlers.ChangePassword) // PUT /api/users/me/password

			// Two-factor authentication
			r.Get("/2fa/status", handlers.Get2FAStatus)                 // GET /api/users/me/2fa/status
			r.Post("/2fa/setup", handlers.Setup2FA)                     // POST /api/users/me/2fa/setup
			r.Post("/2fa/verify", handlers.Verify2FA)                   // POST /api/users/me/2fa/verify
			r.Post("/2fa/disable", handlers.Disable2FA)                 // POST /api/users/me/2fa/disable
			r.Post("/2fa/backup-codes", handlers.RegenerateBackupCodes) // POST /api/users/me/2fa/backup-codes

			// Account deletion
			r.Post("/delete", handlers.RequestAccountDeletion)       // POST /api/users/me/delete
			r.Delete("/delete", handlers.CancelAccountDeletion)      // DELETE /api/users/me/delete - Cancel deletion
			r.Post("/delete/cancel", handlers.CancelAccountDeletion) // POST /api/users/me/delete/cancel (backward compat)

			// Data export
			r.Post("/export", handlers.RequestDataExport)          // POST /api/users/me/export
			r.Get("/export", handlers.GetDataExportStatus)         // GET /api/users/me/export
			r.Get("/export/download", handlers.DownloadDataExport) // GET /api/users/me/export/download

			// Avatar management
			r.Post("/avatar", handlers.UploadAvatar)   // POST /api/users/me/avatar
			r.Delete("/avatar", handlers.DeleteAvatar) // DELETE /api/users/me/avatar

			// Connected accounts (OAuth)
			r.Get("/connected-accounts", handlers.GetConnectedAccounts)            // GET /api/users/me/connected-accounts
			r.Delete("/connected-accounts/{provider}", handlers.DisconnectAccount) // DELETE /api/users/me/connected-accounts/{provider}

			// API keys management
			r.Get("/api-keys", handlers.GetUserAPIKeys)            // GET /api/users/me/api-keys
			r.Post("/api-keys", handlers.CreateUserAPIKey)         // POST /api/users/me/api-keys
			r.Get("/api-keys/{id}", handlers.GetUserAPIKey)        // GET /api/users/me/api-keys/{id}
			r.Put("/api-keys/{id}", handlers.UpdateUserAPIKey)     // PUT /api/users/me/api-keys/{id}
			r.Delete("/api-keys/{id}", handlers.DeleteUserAPIKey)  // DELETE /api/users/me/api-keys/{id}
			r.Post("/api-keys/{id}/test", handlers.TestUserAPIKey) // POST /api/users/me/api-keys/{id}/test
		})

		// Specific user routes
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.GetUser()) // GET /api/users/{id} - Get user by ID

			// Protected operations - require authentication
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Put("/", handlers.UpdateUser()) // PUT /api/users/{id} - Update user (owner or admin)
			})

			// Admin-only operations
			r.Group(func(r chi.Router) {
				r.Use(auth.AdminMiddleware) // Requires admin or super_admin role
				r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))
				r.Delete("/", handlers.DeleteUser())        // DELETE /api/users/{id} - Delete user (admin only)
				r.Patch("/role", handlers.UpdateUserRole()) // PATCH /api/users/{id}/role - Update user role (admin only)
			})
		})
	})

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

	// Billing routes (Stripe)
	r.Route("/billing", func(r chi.Router) {
		r.Use(ratelimit.NewAPIRateLimitMiddleware(rateLimitConfig))

		// Public billing endpoints
		r.Get("/config", stripe.GetBillingConfig()) // GET /api/billing/config - Get publishable key
		r.Get("/plans", stripe.GetPlans())          // GET /api/billing/plans - Get available plans

		// Protected billing endpoints - require authentication
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))

			r.Post("/checkout", stripe.CreateCheckoutSession(stripeConfig)) // POST /api/billing/checkout
			r.Post("/portal", stripe.CreatePortalSession(stripeConfig))     // POST /api/billing/portal
			r.Get("/subscription", stripe.GetSubscription())                // GET /api/billing/subscription
		})
	})

	// Webhook routes (no auth - uses signature verification)
	r.Post("/webhooks/stripe", stripe.HandleWebhook(stripeConfig))

	// Usage metering routes
	r.Route("/usage", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))

		r.Get("/", usageHandler.GetCurrentUsage)        // GET /api/usage - Current period usage
		r.Get("/history", usageHandler.GetUsageHistory) // GET /api/usage/history - Usage history

		// Alerts
		r.Get("/alerts", usageHandler.GetAlerts)                          // GET /api/usage/alerts
		r.Post("/alerts/{id}/acknowledge", usageHandler.AcknowledgeAlert) // POST /api/usage/alerts/{id}/acknowledge
	})

	// AI routes (Gemini) - require authentication with separate rate limit tier
	r.Route("/ai", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Use(ratelimit.NewAIRateLimitMiddleware(rateLimitConfig))

		r.Post("/chat", handlers.AIChat)                  // POST /api/ai/chat - Chat completion
		r.Post("/chat/stream", handlers.AIChatStream)     // POST /api/ai/chat/stream - Streaming chat (SSE)
		r.Post("/chat/advanced", handlers.AIChatAdvanced) // POST /api/ai/chat/advanced - Function calling & JSON mode
		r.Post("/analyze-image", handlers.AIAnalyzeImage) // POST /api/ai/analyze-image - Image analysis
		r.Post("/embeddings", handlers.AIEmbeddings)      // POST /api/ai/embeddings - Generate embeddings
	})

	// Feature flags - public endpoint for current user
	r.Route("/feature-flags", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/", handlers.GetFeatureFlagsForUser) // GET /api/feature-flags - Get flags for current user
	})

	// Admin routes - require admin or super_admin role
	r.Route("/admin", func(r chi.Router) {
		r.Use(auth.AdminMiddleware) // Requires admin or super_admin role (includes AuthMiddleware)

		// Dashboard stats
		r.Get("/stats", handlers.GetAdminStats) // GET /api/admin/stats

		// Audit logs
		r.Get("/audit-logs", handlers.GetAuditLogs) // GET /api/admin/audit-logs

		// User impersonation
		r.Post("/impersonate", handlers.ImpersonateUser)        // POST /api/admin/impersonate
		r.Post("/stop-impersonate", handlers.StopImpersonation) // POST /api/admin/stop-impersonate

		// User management
		r.Get("/users", handlers.SearchUsers)             // GET /api/admin/users?query=... - Search users for command palette
		r.Get("/users/deleted", handlers.GetDeletedUsers) // GET /api/admin/users/deleted - List soft-deleted users
		r.Route("/users/{id}", func(r chi.Router) {
			r.Put("/role", handlers.AdminUpdateUserRole)   // PUT /api/admin/users/{id}/role
			r.Post("/deactivate", handlers.DeactivateUser) // POST /api/admin/users/{id}/deactivate
			r.Post("/reactivate", handlers.ReactivateUser) // POST /api/admin/users/{id}/reactivate
			r.Post("/restore", handlers.RestoreUser)       // POST /api/admin/users/{id}/restore - Restore soft-deleted user

			// User feature flag overrides
			r.Put("/feature-flags/{key}", handlers.SetUserFeatureFlagOverride)       // PUT /api/admin/users/{id}/feature-flags/{key}
			r.Delete("/feature-flags/{key}", handlers.DeleteUserFeatureFlagOverride) // DELETE /api/admin/users/{id}/feature-flags/{key}
		})

		// Feature flags management
		r.Route("/feature-flags", func(r chi.Router) {
			r.Get("/", handlers.GetFeatureFlags)           // GET /api/admin/feature-flags
			r.Post("/", handlers.CreateFeatureFlag)        // POST /api/admin/feature-flags
			r.Put("/{key}", handlers.UpdateFeatureFlag)    // PUT /api/admin/feature-flags/{key}
			r.Delete("/{key}", handlers.DeleteFeatureFlag) // DELETE /api/admin/feature-flags/{key}
		})

		// System settings management
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", handlers.GetAllSettings)                  // GET /api/admin/settings
			r.Get("/{category}", handlers.GetSettingsByCategory) // GET /api/admin/settings/{category}
			r.Put("/{key}", handlers.UpdateSetting)              // PUT /api/admin/settings/{key}

			// Email settings
			r.Get("/email", handlers.GetEmailSettings)        // GET /api/admin/settings/email
			r.Put("/email", handlers.UpdateEmailSettings)     // PUT /api/admin/settings/email
			r.Post("/email/test", handlers.TestEmailSettings) // POST /api/admin/settings/email/test

			// Security settings
			r.Get("/security", handlers.GetSecuritySettings)    // GET /api/admin/settings/security
			r.Put("/security", handlers.UpdateSecuritySettings) // PUT /api/admin/settings/security

			// Site settings
			r.Get("/site", handlers.GetSiteSettings)    // GET /api/admin/settings/site
			r.Put("/site", handlers.UpdateSiteSettings) // PUT /api/admin/settings/site
		})

		// IP blocklist management
		r.Route("/ip-blocklist", func(r chi.Router) {
			r.Get("/", handlers.GetIPBlocklist)   // GET /api/admin/ip-blocklist
			r.Post("/", handlers.BlockIP)         // POST /api/admin/ip-blocklist
			r.Delete("/{id}", handlers.UnblockIP) // DELETE /api/admin/ip-blocklist/{id}
		})

		// Announcements management
		r.Route("/announcements", func(r chi.Router) {
			r.Get("/", handlers.GetAnnouncements)          // GET /api/admin/announcements
			r.Post("/", handlers.CreateAnnouncement)       // POST /api/admin/announcements
			r.Put("/{id}", handlers.UpdateAnnouncement)    // PUT /api/admin/announcements/{id}
			r.Delete("/{id}", handlers.DeleteAnnouncement) // DELETE /api/admin/announcements/{id}
		})

		// Email templates management
		r.Route("/email-templates", func(r chi.Router) {
			r.Get("/", handlers.GetEmailTemplates)                 // GET /api/admin/email-templates
			r.Get("/{id}", handlers.GetEmailTemplate)              // GET /api/admin/email-templates/{id}
			r.Put("/{id}", handlers.UpdateEmailTemplate)           // PUT /api/admin/email-templates/{id}
			r.Post("/{id}/preview", handlers.PreviewEmailTemplate) // POST /api/admin/email-templates/{id}/preview
		})

		// System health monitoring
		r.Route("/health", func(r chi.Router) {
			r.Get("/", handlers.GetSystemHealth)           // GET /api/admin/health
			r.Get("/database", handlers.GetDatabaseHealth) // GET /api/admin/health/database
			r.Get("/cache", handlers.GetCacheHealth)       // GET /api/admin/health/cache
		})
	})

	// Public changelog
	r.Get("/changelog", handlers.GetChangelog) // GET /api/changelog

	// Announcements
	r.Route("/announcements", func(r chi.Router) {
		r.Get("/", handlers.GetActiveAnnouncements) // GET /api/announcements - Get active announcements

		// Authenticated announcement actions
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Post("/{id}/dismiss", handlers.DismissAnnouncement)         // POST /api/announcements/{id}/dismiss
			r.Get("/unread-modals", handlers.GetUnreadModalAnnouncements) // GET /api/announcements/unread-modals
			r.Post("/{id}/read", handlers.MarkAnnouncementRead)           // POST /api/announcements/{id}/read
		})
	})

	// Organization routes (multi-tenancy)
	r.Route("/organizations", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Use(ratelimit.NewAPIRateLimitMiddleware(rateLimitConfig))

		// List user's organizations and create new ones
		r.Get("/", orgHandler.ListOrganizations)   // GET /api/organizations
		r.Post("/", orgHandler.CreateOrganization) // POST /api/organizations

		// Organization-specific routes (require org membership)
		r.Route("/{orgSlug}", func(r chi.Router) {
			r.Use(tenantMiddleware.RequireOrganization)

			r.Get("/", orgHandler.GetOrganization)         // GET /api/organizations/{orgSlug}
			r.Post("/leave", orgHandler.LeaveOrganization) // POST /api/organizations/{orgSlug}/leave

			// Admin+ only routes
			r.Group(func(r chi.Router) {
				r.Use(tenantMiddleware.RequireOrgRole(models.OrgRoleAdmin))
				r.Put("/", orgHandler.UpdateOrganization) // PUT /api/organizations/{orgSlug}

				// Member management
				r.Get("/members", orgHandler.ListMembers)                    // GET /api/organizations/{orgSlug}/members
				r.Post("/members/invite", orgHandler.InviteMember)           // POST /api/organizations/{orgSlug}/members/invite
				r.Put("/members/{userId}/role", orgHandler.UpdateMemberRole) // PUT /api/organizations/{orgSlug}/members/{userId}/role
				r.Delete("/members/{userId}", orgHandler.RemoveMember)       // DELETE /api/organizations/{orgSlug}/members/{userId}

				// Invitation management
				r.Get("/invitations", orgHandler.ListInvitations)                    // GET /api/organizations/{orgSlug}/invitations
				r.Delete("/invitations/{invitationId}", orgHandler.CancelInvitation) // DELETE /api/organizations/{orgSlug}/invitations/{invitationId}

				// Billing (view only for admin+)
				r.Get("/billing", orgHandler.GetOrganizationBilling) // GET /api/organizations/{orgSlug}/billing
			})

			// Owner only routes
			r.Group(func(r chi.Router) {
				r.Use(tenantMiddleware.RequireOrgRole(models.OrgRoleOwner))
				r.Delete("/", orgHandler.DeleteOrganization) // DELETE /api/organizations/{orgSlug}

				// Billing management (owner only)
				r.Post("/billing/checkout", orgHandler.CreateOrganizationCheckout)    // POST /api/organizations/{orgSlug}/billing/checkout
				r.Post("/billing/portal", orgHandler.CreateOrganizationBillingPortal) // POST /api/organizations/{orgSlug}/billing/portal
			})
		})
	})

	// Invitation acceptance (separate from org routes - user may not be a member yet)
	r.Route("/invitations", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/accept", orgHandler.AcceptInvitation) // POST /api/invitations/accept?token=xxx
	})
}
