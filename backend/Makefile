.PHONY: db-up db-down db-reset db-logs db-shell dev clean swagger swagger-check swagger-validate swagger-install \
	test test-v test-race test-cover test-cover-check test-cover-show \
	migrate-install migrate-create migrate-up migrate-down migrate-reset migrate-version migrate-force \
	seed seed-fresh

# Database configuration
DB_NAME=starter_kit_db
DB_USER=devuser
DB_PASS=devpass
DB_PORT=5432
CONTAINER_NAME=postgres-dev

# Start PostgreSQL container
db-up:
	@echo "Starting PostgreSQL container..."
	@docker run --name $(CONTAINER_NAME) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASS) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-d postgres:15
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "PostgreSQL is ready!"

# Stop and remove PostgreSQL container
db-down:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(CONTAINER_NAME) 2>/dev/null || true
	@echo "PostgreSQL container stopped and removed"

# Reset database (stop, remove, start fresh)
db-reset: db-down db-up

# Show database logs
db-logs:
	@docker logs $(CONTAINER_NAME)

# Connect to database shell
db-shell:
	@docker exec -it $(CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME)

# Start development environment (database + air)
dev: db-up
	@echo "Starting development server..."
	@sleep 3
	@air

# Check if swag is installed
swagger-check:
	@which swag > /dev/null 2>&1 || (echo "Error: swag not installed. Run: make swagger-install" && exit 1)

# Generate Swagger documentation
swagger: swagger-check
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/main.go -o docs
	@sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go
	@echo "Swagger documentation generated in docs/"

# Validate Swagger docs are up-to-date
swagger-validate: swagger-check
	@echo "Validating Swagger documentation..."
	@mkdir -p /tmp/swagger-validate
	@swag init -g cmd/main.go -o /tmp/swagger-validate 2>/dev/null
	@sed -i '/LeftDelim:/d; /RightDelim:/d' /tmp/swagger-validate/docs.go 2>/dev/null || true
	@if ! diff -q docs/swagger.json /tmp/swagger-validate/swagger.json > /dev/null 2>&1; then \
		echo "Error: Swagger docs outdated! Run 'make swagger' and commit."; \
		rm -rf /tmp/swagger-validate; \
		exit 1; \
	fi
	@rm -rf /tmp/swagger-validate
	@echo "Swagger documentation is up-to-date."

# Install swag CLI
swagger-install:
	@echo "Installing swag v1.16.6..."
	@go install github.com/swaggo/swag/cmd/swag@v1.16.6
	@echo "Done. Ensure \$$(go env GOPATH)/bin is in PATH."

# Clean up everything
clean: db-down
	@docker image rm postgres:15 2>/dev/null || true
	@echo "Cleaned up Docker images"

# Run all tests
test:
	@go test ./internal/...

# Run tests with verbose output
test-v:
	@go test -v ./internal/...

# Run tests with race detection
test-race:
	@go test -race ./internal/...

# Run tests with coverage report
test-cover:
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Coverage threshold (default 70%)
COVERAGE_THRESHOLD ?= 70

# Run tests with coverage and fail if below threshold
test-cover-check:
	@echo "Running tests with coverage threshold $(COVERAGE_THRESHOLD)%..."
	@go test -coverprofile=coverage.out ./internal/...
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${coverage}%"; \
	if [ $$(echo "$$coverage < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "ERROR: Coverage $${coverage}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "OK: Coverage meets threshold"; \
	fi

# Show current coverage percentage
test-cover-show:
	@go test -coverprofile=coverage.out ./internal/... 2>/dev/null
	@go tool cover -func=coverage.out | grep total

# ============================================
# DATABASE MIGRATIONS (golang-migrate)
# ============================================

MIGRATIONS_PATH=./migrations
DB_URL=postgres://$(DB_USER):$(DB_PASS)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Install golang-migrate CLI
migrate-install:
	@echo "Installing golang-migrate..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "golang-migrate installed successfully"

# Create a new migration (usage: make migrate-create name=add_users_table)
migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo "Migration files created for: $(name)"

# Run all pending migrations
migrate-up:
	@echo "Running migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "Migrations applied successfully"

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "Rolled back 1 migration"

# Rollback all migrations
migrate-reset:
	@echo "Rolling back all migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all
	@echo "All migrations rolled back"

# Show current migration version
migrate-version:
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

# Force set migration version (use carefully!)
# Usage: make migrate-force version=1
migrate-force:
	@if [ -z "$(version)" ]; then echo "Usage: make migrate-force version=N"; exit 1; fi
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)
	@echo "Forced migration version to: $(version)"

# Validate migrations (run up then down then up)
migrate-validate:
	@echo "Validating migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "Migration validation complete"

# ============================================
# SEED DATA
# ============================================

# Seed the database with test data
seed:
	@echo "Seeding database..."
	@go run cmd/seed/main.go
	@echo "Seeding complete"

# Reset database and seed fresh
seed-fresh: migrate-reset migrate-up seed
	@echo "Fresh seed complete"
