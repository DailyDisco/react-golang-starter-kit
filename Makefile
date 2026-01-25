# ============================================
# React + Go Starter Kit - Docker Development
# ============================================

.PHONY: help dev dev-build prod prod-build prod-rebuild prod-status build rebuild down dev-down prod-down clean logs dev-logs prod-logs \
	backend-logs frontend-logs db-logs test test-backend test-frontend format-backend status health setup \
	seed dev-fresh db-reset shell-backend shell-frontend shell-db restart ps tail \
	observability-up observability-down observability-logs grafana-logs prometheus-logs \
	deploy-vercel deploy-vercel-prod deploy-railway configure-features init \
	rollback frontend-build \
	test-db-up test-db-down test-db-reset test-services-up test-services-down test-integration \
	test-backend-coverage test-frontend-coverage test-e2e coverage coverage-html coverage-check test-clean

# ============================================
# Configuration
# ============================================

ENV_FILE := .env

# Database (defaults match docker-compose.yml and .env.example)
DB_USER    ?= devuser
DB_PASSWORD ?= devpass
DB_NAME    ?= starter_kit_db
DB_PORT    ?= 5432

# Ports
FRONTEND_PORT ?= 5193
BACKEND_PORT  ?= 8080

# ============================================
# Compose Commands (DRY)
# ============================================

COMPOSE_FILES_DEV  := -f docker/compose.yml -f docker/compose.dev.yml
COMPOSE_FILES_PROD := -f docker/compose.yml -f docker/compose.prod.yml
COMPOSE_FILES_OBS  := $(COMPOSE_FILES_DEV) -f docker/compose.observability.yml

DC      := docker compose --env-file $(ENV_FILE)
DC_DEV  := $(DC) $(COMPOSE_FILES_DEV)
DC_PROD := $(DC) $(COMPOSE_FILES_PROD)
DC_OBS  := $(DC) $(COMPOSE_FILES_OBS)

# ============================================
# Help
# ============================================

help: ## Show this help message
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ============================================
# Development
# ============================================

dev: ## Start development environment (auto-seeds on startup)
	@$(DC_DEV) up -d
	@echo "Services starting... Run 'make logs' to watch."

dev-build: ## Rebuild and start development environment
	@$(DC_DEV) up -d --build

dev-logs: logs ## Alias for logs

dev-down: ## Stop development environment
	@$(DC_DEV) down

dev-fresh: ## Start dev with fresh database (auto-seeds on startup)
	@$(DC_DEV) down -v
	@$(DC_DEV) up -d
	@echo "Fresh database created. Auto-seeding on backend startup."

restart: ## Restart all development services
	@$(DC_DEV) restart

# ============================================
# Production (Blue-Green Deployment)
# ============================================

prod: ## Deploy with zero downtime (blue-green)
	@./scripts/deploy-bluegreen.sh

prod-build: ## Build production images without deploying
	@$(DC_PROD) build

prod-rebuild: ## Rebuild production images without cache
	@$(DC_PROD) build --no-cache

prod-status: ## Show production deployment status
	@./scripts/deploy-bluegreen.sh --status

prod-logs: ## View production logs
	@$(DC_PROD) logs -f

prod-down: ## Stop production environment
	@$(DC_PROD) down

rollback: ## Rollback to previous environment
	@./scripts/deploy-bluegreen.sh --rollback

frontend-build: ## Build frontend for production deployment
	@cd frontend && npm run build
	@echo "Frontend built to frontend/dist/"
	@echo "Deploy to: Vercel, Cloudflare Pages, S3, or any static host"

# ============================================
# Build
# ============================================

build: ## Build all development services
	@$(DC_DEV) build

rebuild: ## Rebuild all services without cache
	@$(DC_DEV) build --no-cache

# ============================================
# Logs
# ============================================

logs: ## View logs from all services (follow)
	@$(DC_DEV) logs -f

tail: ## View last 100 lines of logs
	@$(DC_DEV) logs --tail=100

backend-logs: ## View backend service logs
	@$(DC_DEV) logs -f backend

frontend-logs: ## View frontend service logs
	@$(DC_DEV) logs -f frontend

db-logs: ## View database logs
	@$(DC_DEV) logs -f postgres

# ============================================
# Shell Access
# ============================================

shell-backend: ## Access backend container shell
	@$(DC_DEV) exec backend sh

shell-frontend: ## Access frontend container shell
	@$(DC_DEV) exec frontend sh

shell-db: ## Access database (psql)
	@$(DC_DEV) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# ============================================
# Database
# ============================================

db-reset: ## Reset database (WARNING: deletes all data)
	@$(DC_DEV) down -v
	@$(DC_DEV) up -d postgres
	@echo "Database reset. Run 'make dev' to start all services."

seed: ## Seed the database with test data
	@$(DC_DEV) exec backend go run ./cmd/seed

# ============================================
# Status & Health
# ============================================

status: ps ## Alias for ps

ps: ## Show status of all services
	@$(DC_DEV) ps

health: ## Check health of all services
	@printf "Backend:  " && curl -sf http://localhost:$(BACKEND_PORT)/health && echo "✓ healthy" || echo "✗ not healthy"
	@printf "Frontend: " && curl -sf http://localhost:$(FRONTEND_PORT) >/dev/null && echo "✓ healthy" || echo "✗ not healthy"
	@printf "Database: " && $(DC_DEV) exec -T postgres pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1 && echo "✓ ready" || echo "✗ not ready"

# ============================================
# Testing
# ============================================

test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend unit tests
	@echo "Running backend tests..."
	@$(DC_DEV) exec -T backend go test ./internal/... 2>/dev/null || (cd backend && go test ./internal/...)

test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	@cd frontend && npm run test:fast

# ============================================
# Testing (Extended)
# ============================================

test-db-up: ## Start test database
	@echo "Starting test database..."
	@docker compose -f docker/compose.test.yml up -d postgres-test dragonfly-test
	@echo "Waiting for database to be ready..."
	@until docker exec react-golang-postgres-test pg_isready -U testuser -d starter_kit_test > /dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "Test database is ready!"

test-db-down: ## Stop test database
	@docker compose -f docker/compose.test.yml down -v

test-db-reset: test-db-down test-db-up ## Reset test database

test-services-up: ## Start all test services (DB, Redis, LocalStack, Mailpit)
	@echo "Starting all test services..."
	@docker compose -f docker/compose.test.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Test services are ready!"

test-services-down: ## Stop all test services
	@docker compose -f docker/compose.test.yml down -v

test-integration: test-db-up ## Run Go integration tests with Docker database
	@echo "Running integration tests..."
	@cd backend && INTEGRATION_TEST=true \
		TEST_DB_HOST=localhost \
		TEST_DB_PORT=5433 \
		TEST_DB_USER=testuser \
		TEST_DB_PASSWORD=testpass \
		TEST_DB_NAME=starter_kit_test \
		go test -v -race -timeout 10m ./internal/...
	@echo "Integration tests complete!"

test-backend-coverage: ## Run backend tests with coverage
	@echo "Running backend tests with coverage..."
	@cd backend && go test -v -race -timeout 5m -coverprofile=coverage.out -covermode=atomic ./internal/...
	@cd backend && go tool cover -func=coverage.out | tail -1

test-frontend-coverage: ## Run frontend tests with coverage
	@echo "Running frontend tests with coverage..."
	@cd frontend && npm run test:coverage

test-e2e: ## Run Playwright E2E tests
	@echo "Running E2E tests..."
	@cd frontend && npm run test:e2e

coverage: test-backend-coverage test-frontend-coverage ## Generate all coverage reports

coverage-html: ## Generate HTML coverage reports
	@echo "Generating backend coverage HTML..."
	@cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Backend coverage report: backend/coverage.html"
	@echo "Frontend coverage report: frontend/coverage/"

coverage-check: ## Check coverage meets 70% threshold
	@echo "Checking coverage thresholds..."
	@cd backend && go test -coverprofile=coverage.out ./internal/... 2>/dev/null && \
		coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Backend coverage: $${coverage}%"; \
		if [ $$(echo "$$coverage < 70" | bc -l) -eq 1 ]; then \
			echo "ERROR: Coverage $${coverage}% is below 70% threshold"; \
			exit 1; \
		fi
	@echo "Coverage check passed!"

test-clean: ## Clean test artifacts
	@rm -f backend/coverage.out backend/coverage.html backend/coverage-integration.out
	@rm -rf frontend/coverage frontend/playwright-report frontend/playwright-results
	@echo "Test artifacts cleaned"

# ============================================
# Code Quality
# ============================================

format-backend: ## Format backend Go code
	@cd backend && go fmt ./...

sync-deps: ## Sync all dependencies (npm install + go mod tidy)
	@echo "Syncing frontend dependencies..."
	@bash -c 'export NVM_DIR="$$HOME/.nvm"; [ -s "$$NVM_DIR/nvm.sh" ] && . "$$NVM_DIR/nvm.sh"; cd frontend && npm install'
	@echo "Syncing backend dependencies..."
	@cd backend && go mod tidy
	@echo "Dependencies synced!"

sync-frontend: ## Sync frontend dependencies only
	@bash -c 'export NVM_DIR="$$HOME/.nvm"; [ -s "$$NVM_DIR/nvm.sh" ] && . "$$NVM_DIR/nvm.sh"; cd frontend && npm install'

sync-backend: ## Sync backend dependencies only
	@cd backend && go mod tidy

# ============================================
# Cleanup
# ============================================

down: ## Stop all running containers
	@$(DC_DEV) down
	@$(DC_PROD) down 2>/dev/null || true

clean: ## Clean up containers, volumes, and images
	@$(DC_DEV) down -v --remove-orphans
	@$(DC_PROD) down -v --remove-orphans 2>/dev/null || true
	@docker system prune -f

# ============================================
# Setup
# ============================================

setup: ## Initial setup - copy env file
	@if [ -f .env ]; then \
		echo "Warning: .env already exists. Remove it first to reset."; \
	else \
		cp .env.example .env && \
		echo "Created .env from .env.example" && \
		echo "Customize as needed, then run 'make dev'."; \
	fi

# ============================================
# Observability
# ============================================

observability-up: ## Start observability stack (Prometheus + Grafana)
	@docker network create react-golang-starter-network 2>/dev/null || true
	@$(DC_OBS) up -d prometheus grafana
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana:    http://localhost:3001 (admin/admin)"

observability-down: ## Stop observability stack
	@$(DC_OBS) stop prometheus grafana

observability-logs: ## View observability stack logs
	@$(DC_OBS) logs -f prometheus grafana

grafana-logs: ## View Grafana logs
	@$(DC_OBS) logs -f grafana

prometheus-logs: ## View Prometheus logs
	@$(DC_OBS) logs -f prometheus

# ============================================
# External Deployment
# ============================================

deploy-vercel: ## Deploy frontend to Vercel
	@./scripts/deploy-vercel.sh

deploy-vercel-prod: ## Deploy frontend to Vercel (production)
	@./scripts/deploy-vercel.sh --prod

deploy-railway: ## Deploy backend to Railway
	@./scripts/deploy-railway.sh

# ============================================
# Project Setup
# ============================================

configure-features: ## Interactive feature configuration wizard
	@./scripts/configure-features.sh

init: ## Initialize a new project from this template
	@./init-project.sh

