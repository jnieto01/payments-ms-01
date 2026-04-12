.PHONY: help build run test clean docker-build docker-run docker-stop migrate-up migrate-down migrate-create migrate-version migrate-force migrate-reset

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/user-ms cmd/api/main.go

run: ## Run the application
	go run cmd/api/main.go

seed-mp-plans: ## Create MercadoPago preapproval plans for active local plans without one (idempotent)
	go run ./cmd/seed/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

docker-build: ## Build Docker image
	docker build -t user-ms-go:latest .

docker-run: ## Run with Docker Compose
	docker-compose up -d

docker-stop: ## Stop Docker Compose
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f user-ms

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

# Database Migration Commands
migrate-up: ## Run all pending migrations
	go run cmd/migrate/main.go -command=up

migrate-down: ## Rollback the last migration
	go run cmd/migrate/main.go -command=down

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name. Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	go run cmd/migrate/main.go -command=create $(name)

migrate-version: ## Show current migration version
	go run cmd/migrate/main.go -command=version

migrate-force: ## Force migration to specific version (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "Error: Please provide a version. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	go run cmd/migrate/main.go -command=force -version=$(version)

migrate-steps: ## Run N migration steps (usage: make migrate-steps steps=2 or steps=-2)
	@if [ -z "$(steps)" ]; then \
		echo "Error: Please provide steps. Usage: make migrate-steps steps=2"; \
		exit 1; \
	fi
	go run cmd/migrate/main.go -command=steps -steps=$(steps)

migrate-reset: ## Reset migrations (drops schema_migrations table)
	@echo "⚠️  WARNING: This will drop the schema_migrations table!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read confirm
	go run cmd/migrate-reset/main.go
