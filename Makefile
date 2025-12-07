# Makefile for Mining Finance System Backend

# Default values
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_USER ?= postgres
DB_PASSWORD ?= miningFinanceDB_2025
DB_NAME ?= mining_data
JWT_SECRET ?= prod-mining-finance-secret-2025
PORT ?= 9006

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
NC := \033[0m # No Color

.PHONY: help start stop build clean test deps docker-up docker-down logs

# Default target
help: ## Show this help message
	@echo "$(GREEN)Mining Finance System Backend$(NC)"
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

start: build ## Start the backend server using pre-built binary
	@echo "$(GREEN)Starting backend server...$(NC)"
	@echo "Database: $(DB_HOST):$(DB_PORT)"
	@echo "Port: $(PORT)"
	@echo "Database: $(DB_NAME)"
	@echo ""
	# Ensure Docker is running and start postgres container if available
	@docker ps >/dev/null 2>&1 && docker-compose up -d postgres || true
	# Wait for Postgres to be ready on the configured port
	@echo "Waiting for Postgres on $(DB_HOST):$(DB_PORT)..."
	@i=0; until nc -z $(DB_HOST) $(DB_PORT) >/dev/null 2>&1; do \
	  i=$$((i+1)); \
	  if [ $$i -gt 30 ]; then echo "$(RED)Postgres not reachable on $(DB_HOST):$(DB_PORT)$(NC)"; exit 1; fi; \
	  sleep 1; \
	done; \
	echo "$(GREEN)Postgres is up$(NC)";
	@echo "$(GREEN)Starting binary...$(NC)"
	@DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) JWT_SECRET=$(JWT_SECRET) PORT=$(PORT) ./bin/api

start-bg: build ## Start the backend server in background using pre-built binary
	@echo "$(GREEN)Starting backend server in background...$(NC)"
	@echo "Database: $(DB_HOST):$(DB_PORT)"
	@echo "Port: $(PORT)"
	@echo "Database: $(DB_NAME)"
	@echo ""
	@echo "$(GREEN)Starting binary in background...$(NC)"
	@DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) JWT_SECRET=$(JWT_SECRET) PORT=$(PORT) ./bin/api &

stop: ## Stop the backend server
	@echo "$(RED)Stopping backend server...$(NC)"
	@pkill -f "./bin/api" || pkill -f "bin/api" || echo "No backend process found"

restart: stop start ## Restart the backend server

build: ## Build the backend binary
	@echo "$(GREEN)Building backend binary...$(NC)"
	@mkdir -p bin
	@go build -ldflags="-s -w" -o bin/api cmd/api/main.go cmd/api/config.go cmd/api/db.go
	@chmod +x bin/api
	@echo "$(GREEN)Binary built: bin/api$(NC)"

run-binary: build ## Build and run the binary
	@echo "$(GREEN)Running binary...$(NC)"
	@DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) JWT_SECRET=$(JWT_SECRET) PORT=$(PORT) ./bin/api

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@go clean

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test ./...

deps: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	@go mod tidy
	@go mod download

docker-up: ## Start PostgreSQL and Redis containers
	@echo "$(GREEN)Starting Docker containers...$(NC)"
	@docker-compose up -d postgres redis
	@echo "$(GREEN)Docker containers started$(NC)"

docker-down: ## Stop Docker containers
	@echo "$(RED)Stopping Docker containers...$(NC)"
	@docker-compose down

docker-logs: ## Show Docker container logs
	@docker-compose logs -f

logs: ## Show backend logs (if running in background)
	@echo "$(GREEN)Backend logs:$(NC)"
	@ps aux | grep -E "(bin/api|cmd/api)" | grep -v grep || echo "Backend not running"

dev: docker-up start ## Start development environment (Docker + Backend)

# Health check
health: ## Check if backend is running
	@echo "$(GREEN)Checking backend health...$(NC)"
	@curl -s http://localhost:$(PORT)/health && echo "$(GREEN)✓ Backend is healthy$(NC)" || echo "$(RED)✗ Backend is not responding$(NC)"

# Test endpoints
test-signup: ## Test signup endpoint
	@echo "$(GREEN)Testing signup endpoint...$(NC)"
	@curl -X POST http://localhost:$(PORT)/api/v1/auth/signup \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","name":"Test User","password":"password123","phone":"+256700123456"}' \
		| jq . || echo "Response received (jq not available)"

test-login: ## Test login endpoint
	@echo "$(GREEN)Testing login endpoint...$(NC)"
	@curl -X POST http://localhost:$(PORT)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","password":"password123"}' \
		| jq . || echo "Response received (jq not available)"

# Database operations
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "$(RED)WARNING: This will delete all data in the database!$(NC)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker-compose down -v
	@docker-compose up -d postgres
	@echo "$(GREEN)Database reset complete$(NC)"

# Development helpers
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@echo "$(GREEN)Tools installed$(NC)"

# Hot reload (requires air)
dev-reload: ## Start with hot reload (requires air)
	@echo "$(GREEN)Starting with hot reload...$(NC)"
	@air

# Environment setup
setup: deps docker-up ## Setup development environment
	@echo "$(GREEN)Development environment setup complete!$(NC)"
	@echo "Run 'make start' to start the backend"
