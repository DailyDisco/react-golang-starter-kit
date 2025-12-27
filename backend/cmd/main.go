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

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/email"
	"react-golang-starter/internal/handlers"
	"react-golang-starter/internal/jobs"
	"react-golang-starter/internal/middleware"
	"react-golang-starter/internal/ratelimit"
	"react-golang-starter/internal/services"
	"react-golang-starter/internal/stripe"

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
	// Load environment variables - try multiple locations
	err := godotenv.Load(".env.local")
	if err != nil {
		// Try .env as fallback (for when running from backend/ directory)
		err = godotenv.Load(".env")
		if err != nil {
			zerologlog.Info().Msg("No .env.local or .env file found, using system environment variables")
		}
	}

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

	// Create Chi router
	r := chi.NewRouter()

	// Global middleware
	// Compression middleware for improved performance (must be first)
	r.Use(chimiddleware.Compress(5, "application/json", "text/plain", "text/html"))

	// Request ID middleware (before logger so IDs are logged)
	r.Use(middleware.RequestIDMiddleware)

	// Sentry middleware (captures panics and reports errors)
	r.Use(middleware.SentryMiddleware(sentryConfig))

	r.Use(middleware.StructuredLoggerWithConfig(logConfig))
	r.Use(chimiddleware.Recoverer)

	// CORS middleware for React frontend (MUST be before rate limiting so preflight OPTIONS requests work)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   getAllowedOrigins(),
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

	// Routes
	setupRoutes(r, rateLimitConfig, stripeConfig, appService, fileService)

	// Create server
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start job processing in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

		// Stop job processing first
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

func setupRoutes(r chi.Router, rateLimitConfig *ratelimit.Config, stripeConfig *stripe.Config, appService *handlers.Service, fileService *services.FileService) {
	// Simple test route at root level
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte("Test route working!")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
	})

	// API routes setup - shared between /api and /api/v1
	apiRoutes := func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("API test route working!")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		})
		r.Get("/health", appService.HealthCheck)
		setupAPIRoutes(r, rateLimitConfig, stripeConfig, appService, fileService)
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
}

// setupAPIRoutes configures all API endpoints
func setupAPIRoutes(r chi.Router, rateLimitConfig *ratelimit.Config, stripeConfig *stripe.Config, appService *handlers.Service, fileService *services.FileService) {
	_ = appService // Avoid unused variable warning

	// CSRF token endpoint - allows frontend to get a fresh CSRF token
	csrfConfig := middleware.LoadCSRFConfig()
	r.Get("/csrf-token", middleware.GetCSRFToken(csrfConfig))

	// Authentication routes
	r.Route("/auth", func(r chi.Router) {
		r.Use(ratelimit.NewAuthRateLimitMiddleware(rateLimitConfig))

		// Public auth endpoints
		r.Post("/register", auth.RegisterUser)      // POST /api/auth/register
		r.Post("/login", auth.LoginUser)            // POST /api/auth/login
		r.Post("/logout", auth.LogoutUser)          // POST /api/auth/logout
		r.Post("/refresh", auth.RefreshAccessToken) // POST /api/auth/refresh - Exchange refresh token for new access token

		// Protected auth endpoints
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Get("/me", auth.GetCurrentUser) // GET /api/auth/me
		})

		// Password reset (public)
		r.Post("/reset-password", auth.RequestPasswordReset)  // POST /api/auth/reset-password
		r.Post("/reset-password/confirm", auth.ResetPassword) // POST /api/auth/reset-password/confirm
		r.Get("/verify-email", auth.VerifyEmail)              // GET /api/auth/verify-email

		// OAuth routes
		r.Route("/oauth", func(r chi.Router) {
			// Public OAuth endpoints
			r.Get("/{provider}", auth.GetOAuthURL)                  // GET /api/auth/oauth/{provider} - Get OAuth URL
			r.Get("/{provider}/callback", auth.HandleOAuthCallback) // GET /api/auth/oauth/{provider}/callback - OAuth callback

			// Protected OAuth endpoints
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
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

		// Protected routes - require authentication
		r.Route("/", func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Use(ratelimit.NewUserRateLimitMiddleware(rateLimitConfig))

			r.Get("/me", handlers.GetCurrentUser())    // GET /api/users/me - Get current user
			r.Put("/me", handlers.UpdateCurrentUser()) // PUT /api/users/me - Update current user

			// User preferences
			r.Get("/me/preferences", handlers.GetUserPreferences)    // GET /api/users/me/preferences
			r.Put("/me/preferences", handlers.UpdateUserPreferences) // PUT /api/users/me/preferences

			// Session management
			r.Get("/me/sessions", handlers.GetUserSessions)       // GET /api/users/me/sessions
			r.Delete("/me/sessions", handlers.RevokeAllSessions)  // DELETE /api/users/me/sessions - Revoke all other sessions
			r.Delete("/me/sessions/{id}", handlers.RevokeSession) // DELETE /api/users/me/sessions/{id}

			// Login history
			r.Get("/me/login-history", handlers.GetLoginHistory) // GET /api/users/me/login-history

			// Password change
			r.Put("/me/password", handlers.ChangePassword) // PUT /api/users/me/password

			// Two-factor authentication
			r.Get("/me/2fa/status", handlers.Get2FAStatus)                 // GET /api/users/me/2fa/status
			r.Post("/me/2fa/setup", handlers.Setup2FA)                     // POST /api/users/me/2fa/setup
			r.Post("/me/2fa/verify", handlers.Verify2FA)                   // POST /api/users/me/2fa/verify
			r.Post("/me/2fa/disable", handlers.Disable2FA)                 // POST /api/users/me/2fa/disable
			r.Post("/me/2fa/backup-codes", handlers.RegenerateBackupCodes) // POST /api/users/me/2fa/backup-codes

			// Account deletion
			r.Post("/me/delete", handlers.RequestAccountDeletion)       // POST /api/users/me/delete
			r.Post("/me/delete/cancel", handlers.CancelAccountDeletion) // POST /api/users/me/delete/cancel

			// Data export
			r.Post("/me/export", handlers.RequestDataExport) // POST /api/users/me/export
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

	// Public announcements
	r.Route("/announcements", func(r chi.Router) {
		r.Get("/", handlers.GetActiveAnnouncements) // GET /api/announcements - Get active announcements

		// Authenticated users can dismiss announcements
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Post("/{id}/dismiss", handlers.DismissAnnouncement) // POST /api/announcements/{id}/dismiss
		})
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
		"http://localhost:5193",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
}
