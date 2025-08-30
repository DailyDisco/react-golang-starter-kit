# React + Go Backend Starter Kit

A modern, production-ready backend built with Go, Chi router, GORM, and PostgreSQL, designed to work seamlessly with a React frontend.

## 🚀 Quick Start

```bash
# Clone and navigate to backend
cd backend

# Start the development environment (PostgreSQL + Go server)
make dev

# The server will be running on http://localhost:8080
# API endpoints available at http://localhost:8080/api/*
```

## 📁 Project Structure

```
backend/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/                   # Private application code
│   ├── database/
│   │   └── database.go         # Database connection & configuration
│   ├── handlers/
│   │   └── handlers.go         # HTTP request handlers (controllers)
│   └── models/
│       └── models.go           # Data models (database schemas)
├── scripts/                    # Utility scripts (optional)
├── .air.toml                   # Hot reload configuration
├── .env                        # Environment variables
├── docker-compose.yml          # Docker services configuration
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── Makefile                    # Development commands
└── README.md                   # This file
```

### Directory Structure Reasoning

#### `/cmd`

Contains application entry points. This is a Go convention where `main.go` files live. Separating this from business logic makes the codebase more maintainable and testable.

#### `/internal`

Private application code that cannot be imported by external applications. This enforces boundaries and prevents external packages from importing internal business logic.

- **`/internal/database`**: Database connection logic, migrations, and database-related utilities
- **`/internal/handlers`**: HTTP handlers (similar to controllers in MVC). Each handler processes HTTP requests and returns responses
- **`/internal/models`**: Data models that define the structure of your database tables using GORM tags

This structure follows Go's [Standard Project Layout](https://github.com/golang-standards/project-layout) and makes the code easy to navigate for any Go developer.

## 🛠 Technology Stack

### Core Framework & Libraries

- **[Go](https://golang.org/)** - High-performance backend language
- **[Chi Router](https://github.com/go-chi/chi)** - Lightweight, idiomatic HTTP router
  - _Why Chi?_ Fast, follows standard `net/http` patterns, composable middleware, excellent for RESTful APIs
- **[GORM](https://gorm.io/)** - Go ORM library
  - _Why GORM?_ Developer-friendly, auto-migration, relationship handling, works with multiple databases
- **[PostgreSQL](https://postgresql.org/)** - Production-ready relational database
  - _Why PostgreSQL?_ ACID compliance, JSON support, excellent performance, widely adopted

### Development Tools

- **[Air](https://github.com/air-verse/air)** - Live reload for Go applications
- **[Docker](https://docker.com/)** - Containerized PostgreSQL for consistent development environment
- **[Make](https://www.gnu.org/software/make/)** - Build automation and common commands

## 🚀 Development Setup

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- Make (usually pre-installed on Linux/macOS)

### Installation

1. **Clone the repository**

   ```bash
   git clone <your-repo>
   cd backend
   ```

2. **Install Go dependencies**

   ```bash
   go mod tidy
   ```

3. **Install Air for hot reload**

   ```bash
   go install github.com/air-verse/air@latest
   ```

4. **Start development environment**
   ```bash
   make dev
   ```

## 🐳 Database Management

### Using Make Commands (Recommended)

```bash
# Start PostgreSQL container
make db-up

# Stop PostgreSQL container
make db-down

# Reset database (fresh start with latest schema)
make db-reset

# View database logs
make db-logs

# Connect to database shell
make db-shell

# Start full development environment (DB + Go server)
make dev

# Clean up everything
make clean
```

### Using Docker Compose (Alternative)

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Reset with fresh data
docker-compose down -v && docker-compose up -d
```

### Database Configuration

The application uses these default credentials (defined in `Makefile` and `.env`):

- **Host:** localhost
- **Port:** 5433 (to avoid conflicts with system PostgreSQL)
- **Database:** devdb
- **Username:** devuser
- **Password:** devpass

You can override these by setting environment variables or modifying the `.env` file.

## 🌐 API Endpoints

### Health Check

```
GET /api/health
```

### Users (CRUD Operations)

```
GET    /api/users     # Get all users
POST   /api/users     # Create new user
GET    /api/users/{id} # Get user by ID
PUT    /api/users/{id} # Update user
DELETE /api/users/{id} # Delete user
```

### Example API Usage

```bash
# Health check
curl http://localhost:8080/api/health

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# Get all users
curl http://localhost:8080/api/users

# Get specific user
curl http://localhost:8080/api/users/1
```

## 🔧 Configuration

### Environment Variables

Create or modify `.env` file:

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=devuser
DB_PASSWORD=devpass
DB_NAME=devdb
DB_SSLMODE=disable
```

### CORS Configuration

The server is pre-configured to allow requests from `http://localhost:3000` (React dev server). Modify in `cmd/main.go` if needed:

```go
cors.Handler(cors.Options{
    AllowedOrigins: []string{"http://localhost:3000"},
    // ... other options
})
```

## 🏗 Adding New Features

### 1. Adding a New Model

```go
// internal/models/models.go
type Product struct {
    gorm.Model
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
}
```

Don't forget to add it to auto-migration in `internal/database/database.go`:

```go
DB.AutoMigrate(&models.User{}, &models.Product{})
```

### 2. Adding New Routes

```go
// cmd/main.go - in setupRoutes function
r.Route("/products", func(r chi.Router) {
    r.Get("/", handlers.GetProducts)
    r.Post("/", handlers.CreateProduct)
    // ... more routes
})
```

### 3. Adding New Handlers

```go
// internal/handlers/handlers.go
func GetProducts(w http.ResponseWriter, r *http.Request) {
    var products []models.Product
    database.DB.Find(&products)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/handlers
```

## 📊 Database Migrations

GORM handles migrations automatically via `AutoMigrate()`. When you:

1. Add new fields to models
2. Create new models
3. Modify existing field types

Simply restart the application or run `make db-reset` for a fresh database.

For production, consider using proper migration tools like [golang-migrate](https://github.com/golang-migrate/migrate).

## 🚀 Deployment

### Railway Deployment

This application is configured to work with Railway out of the box:

1. **PostgreSQL**: Railway provides PostgreSQL automatically
2. **Redis**: By default, Redis is optional (can be disabled)
3. **Environment Variables**: Set the following in Railway environment variables:

```env
# Redis Configuration (Optional)
REDIS_REQUIRED=false

# JWT Secret (Required)
JWT_SECRET=your-secure-jwt-secret-here

# CORS Origins (Update with your frontend URL)
CORS_ALLOWED_ORIGINS=https://your-frontend-app.vercel.app

# Logging
LOG_LEVEL=info
```

### Redis Configuration

The application supports Redis for caching but can run without it:

- **Default**: Redis is required (will fail to start if Redis is unavailable)
- **Optional**: Set `REDIS_REQUIRED=false` to run without Redis
- **Railway**: Redis is not available by default, so set `REDIS_REQUIRED=false`

When Redis is unavailable, the application will:

- Skip all caching operations
- Log warnings about Redis being unavailable
- Continue to function normally using only the database

## 🚨 Troubleshooting

### Common Issues

1. **Port 5432 already in use**

   - The Makefile uses port 5433 to avoid conflicts
   - Check if system PostgreSQL is running: `sudo systemctl status postgresql`

2. **Database connection failed**

   - Ensure Docker container is running: `docker ps`
   - Check container logs: `make db-logs`
   - Verify credentials in `.env` match `Makefile`

3. **Redis connection failed**

   - For Railway deployment: Set `REDIS_REQUIRED=false`
   - For local development: Ensure Redis container is running
   - The application will continue without caching if Redis is unavailable

4. **Module import errors**

   - Ensure `go.mod` module name matches import paths
   - Run `go mod tidy` to clean up dependencies

5. **Air not working**
   - Install Air: `go install github.com/air-verse/air@latest`
   - Add `$HOME/go/bin` to your PATH
   - Alternative: use `go run cmd/main.go`

### Debug Mode

Enable detailed database logging by modifying `internal/database/database.go`:

```go
DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // Add this line
})
```

## 🔒 Security Considerations

This starter kit is configured for development. For production:

1. **Environment Variables**: Use proper secret management
2. **CORS**: Restrict to your production domain
3. **Rate Limiting**: Add rate limiting middleware
4. **Authentication**: Implement JWT or session-based auth
5. **Input Validation**: Add request validation middleware
6. **HTTPS**: Use TLS certificates
7. **Database**: Use connection pooling and read replicas

## 📚 Additional Resources

- [Chi Router Documentation](https://github.com/go-chi/chi)
- [GORM Documentation](https://gorm.io/docs/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [PostgreSQL Documentation](https://postgresql.org/docs/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Add tests if applicable
5. Run tests: `go test ./...`
6. Commit changes: `git commit -m "Add feature"`
7. Push to branch: `git push origin feature-name`
8. Create a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.
