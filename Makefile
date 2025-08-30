# ============================================
# React + Go Starter Kit - Docker Development
# ============================================

.PHONY: help dev prod staging build rebuild stop clean logs shell backend-logs frontend-logs db-logs

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
dev: ## Start development environment with hot reload
	docker compose up -d

dev-logs: ## View development logs
	docker compose logs -f

dev-stop: ## Stop development environment
	docker compose down

# Production commands
prod: ## Start production environment
	docker compose -f docker-compose.prod.yml up -d

prod-logs: ## View production logs
	docker compose -f docker-compose.prod.yml logs -f

prod-stop: ## Stop production environment
	docker compose -f docker-compose.prod.yml down

# Staging commands
staging: ## Start staging environment
	docker compose -f docker-compose.staging.yml up -d

staging-logs: ## View staging logs
	docker compose -f docker-compose.staging.yml logs -f

staging-stop: ## Stop staging environment
	docker compose -f docker-compose.staging.yml down

# Build commands
build: ## Build all services
	docker compose build

rebuild: ## Rebuild all services without cache
	docker compose build --no-cache

prod-build: ## Build production images
	docker compose -f docker-compose.prod.yml build

staging-build: ## Build staging images
	docker compose -f docker-compose.staging.yml build

# Utility commands
stop: ## Stop all running containers
	docker compose down
	docker compose -f docker-compose.prod.yml down
	docker compose -f docker-compose.staging.yml down

clean: ## Clean up containers, volumes, and images
	docker compose down -v
	docker compose -f docker-compose.prod.yml down -v
	docker compose -f docker-compose.staging.yml down -v
	docker system prune -f

logs: ## View logs from all services
	docker compose logs -f

# Service-specific logs
backend-logs: ## View backend service logs
	docker compose logs -f backend

frontend-logs: ## View frontend service logs
	docker compose logs -f frontend

db-logs: ## View database logs
	docker compose logs -f postgres

# Shell access
shell-backend: ## Access backend container shell
	docker compose exec backend sh

shell-frontend: ## Access frontend container shell
	docker compose exec frontend sh

shell-db: ## Access database container shell
	docker compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# Database operations
db-reset: ## Reset database (WARNING: This will delete all data)
	docker compose down -v
	docker compose up -d postgres
	@echo "Database reset complete. Run 'make dev' to start all services."

# Environment setup
setup: ## Initial setup - copy env file and start services
	cp .env.example .env
	@echo "Please edit .env file with your configuration, then run 'make dev'"

# Health checks
health: ## Check health of all services
	@echo "Checking backend health..."
	curl -f http://localhost:8080/health || echo "Backend not healthy"
	@echo "Checking frontend health..."
	curl -f http://localhost:5173 || echo "Frontend not healthy"
	@echo "Checking database..."
	docker compose exec postgres pg_isready -U $(DB_USER) -d $(DB_NAME) || echo "Database not ready"
