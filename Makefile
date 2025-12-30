# ============================================
# React + Go Starter Kit - Docker Development
# ============================================

.PHONY: help dev prod build rebuild stop clean logs shell backend-logs frontend-logs db-logs format-backend observability-up observability-down observability-logs grafana-logs prometheus-logs

# Environment file
ENV_FILE := .env.local

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
dev: ## Start development environment with hot reload and seed data
	docker compose --env-file $(ENV_FILE) up -d
	@echo "Waiting for backend to be ready..."
	@sleep 8
	@docker compose --env-file $(ENV_FILE) exec -T backend go run ./cmd/seed 2>/dev/null || echo "Seeding skipped (backend not ready or already seeded)"

dev-logs: ## View development logs
	docker compose --env-file $(ENV_FILE) logs -f

dev-stop: ## Stop development environment
	docker compose --env-file $(ENV_FILE) down

# Production commands
prod: ## Start production environment
	docker compose --env-file $(ENV_FILE) -f docker-compose.prod.yml up -d

prod-logs: ## View production logs
	docker compose --env-file $(ENV_FILE) -f docker-compose.prod.yml logs -f

prod-stop: ## Stop production environment
	docker compose --env-file $(ENV_FILE) -f docker-compose.prod.yml down

# Build commands
build: ## Build all services
	docker compose --env-file $(ENV_FILE) build

rebuild: ## Rebuild all services without cache
	docker compose --env-file $(ENV_FILE) build --no-cache

prod-build: ## Build production images
	docker compose --env-file $(ENV_FILE) -f docker-compose.prod.yml build

# Utility commands
stop: ## Stop all running containers
	docker compose --env-file $(ENV_FILE) down
	docker compose --env-file $(ENV_FILE) -f docker-compose.prod.yml down

clean: ## Clean up containers, volumes, and images
	docker compose --env-file $(ENV_FILE) down -v
	docker compose --env-file $(ENV_FILE) -f docker-compose.prod.yml down -v
	docker system prune -f

logs: ## View logs from all services
	docker compose --env-file $(ENV_FILE) logs -f

# Service-specific logs
backend-logs: ## View backend service logs
	docker compose --env-file $(ENV_FILE) logs -f backend

frontend-logs: ## View frontend service logs
	docker compose --env-file $(ENV_FILE) logs -f frontend

db-logs: ## View database logs
	docker compose --env-file $(ENV_FILE) logs -f postgres

# Shell access
shell-backend: ## Access backend container shell
	docker compose --env-file $(ENV_FILE) exec backend sh

shell-frontend: ## Access frontend container shell
	docker compose --env-file $(ENV_FILE) exec frontend sh

shell-db: ## Access database container shell
	docker compose --env-file $(ENV_FILE) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# Database operations
db-reset: ## Reset database (WARNING: This will delete all data)
	docker compose --env-file $(ENV_FILE) down -v
	docker compose --env-file $(ENV_FILE) up -d postgres
	@echo "Database reset complete. Run 'make dev' to start all services."

seed: ## Seed the database with test data
	docker compose --env-file $(ENV_FILE) exec backend go run ./cmd/seed

dev-fresh: ## Start dev with fresh database and seed data
	docker compose --env-file $(ENV_FILE) down -v
	docker compose --env-file $(ENV_FILE) up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	docker compose --env-file $(ENV_FILE) exec backend go run ./cmd/seed

# Environment setup
setup: ## Initial setup - copy env file and start services
	cp .env.example .env.local
	@echo "Please edit .env.local file with your configuration, then run 'make dev'"

# Health checks
health: ## Check health of all services
	@echo "Checking backend health..."
	curl -f http://localhost:8080/health || echo "Backend not healthy"
	@echo "Checking frontend health..."
	curl -f http://localhost:5173 || echo "Frontend not healthy"
	@echo "Checking database..."
	docker compose --env-file $(ENV_FILE) exec postgres pg_isready -U $(DB_USER) -d $(DB_NAME) || echo "Database not ready"

# Code formatting
format-backend: ## Format backend Go code
	cd backend && go fmt ./...

# Observability commands
observability-up: ## Start observability stack (Prometheus + Grafana)
	docker network create app-network 2>/dev/null || true
	docker compose --env-file $(ENV_FILE) -f docker-compose.yml -f docker-compose.observability.yml up -d prometheus grafana
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3001 (admin/admin)"

observability-down: ## Stop observability stack
	docker compose --env-file $(ENV_FILE) -f docker-compose.observability.yml down

observability-logs: ## View observability stack logs
	docker compose --env-file $(ENV_FILE) -f docker-compose.observability.yml logs -f

grafana-logs: ## View Grafana logs
	docker compose --env-file $(ENV_FILE) -f docker-compose.observability.yml logs -f grafana

prometheus-logs: ## View Prometheus logs
	docker compose --env-file $(ENV_FILE) -f docker-compose.observability.yml logs -f prometheus
