.PHONY: help migrate db-up db-down generate dev build test clean

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Database operations
db-up: ## Start PostgreSQL database
	docker compose up -d

db-down: ## Stop PostgreSQL database  
	docker compose down

create-migration:
	atlas migrate diff --env local

migrate: ## Apply database migrations
	atlas migrate apply --env local

# Code generation
generate: ## Generate SQLC code
	sqlc generate

# Development
dev: ## Start development server with hot reload
	air

setup: db-up migrate generate ## Setup development environment (database + migrations + code generation)

# Build and test
build: ## Build the application
	go build -o bin/main infra/main.go

test: ## Run tests
	go test ./... -cover

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/

# Complete development workflow
dev-setup: db-up migrate generate dev ## Full development setup and start server