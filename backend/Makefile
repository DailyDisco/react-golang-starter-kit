.PHONY: db-up db-down db-reset db-logs db-shell dev clean swagger

# Database configuration
DB_NAME=devdb
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

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/main.go -o docs
	@sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go
	@echo "Swagger documentation generated in docs/"

# Clean up everything
clean: db-down
	@docker image rm postgres:15 2>/dev/null || true
	@echo "Cleaned up Docker images"
