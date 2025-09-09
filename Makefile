.PHONY: build test run-demo setup-db migrate clean dev full-demo format check-format vet tidy test-coverage quality pre-commit run-api run-scheduler run-worker

# Load environment variables from .env
include .env
export $(shell sed 's/=.*//' .env)

# Docker Compose settings
COMPOSE_FILE=docker-compose.yml
DB_SERVICE=postgres

# Build all binaries
build:
	go build -o bin/aster-api ./cmd/aster-api
	go build -o bin/aster-scheduler ./cmd/aster-scheduler
	go build -o bin/aster-worker ./cmd/aster-worker

# Database setup
setup-db:
	docker compose -f $(COMPOSE_FILE) up -d $(DB_SERVICE)
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker compose -f $(COMPOSE_FILE) exec -T $(DB_SERVICE) pg_isready -U $(POSTGRES_USER); do \
		echo "PostgreSQL is not ready yet..."; \
		sleep 2; \
	done
	@echo "PostgreSQL is ready!"

# Run migrations inside the Postgres container
migrate: setup-db
	@echo "Creating tables..."
	docker compose -f $(COMPOSE_FILE) exec -T $(DB_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f /migrations/001_jobs_table.sql
	docker compose -f $(COMPOSE_FILE) exec -T $(DB_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f /migrations/002_runs_table.sql

# Run all tests
test: migrate
	go test -v ./internal/config
	go test -v ./internal/db
	go test -v ./internal/db/store
	go test -v ./internal/scheduler
	go test -v ./internal/executor
	go test -v ./internal/worker

# Run the simple demo
run-demo: migrate build
	go run ./cmd/demo

# Full system build with all components
full-demo: migrate build
	chmod +x scripts/build.sh
	./scripts/build.sh

# Individual service commands
run-api: migrate
	docker compose -f $(COMPOSE_FILE) up -d postgres
	docker compose -f $(COMPOSE_FILE) up aster-api

run-scheduler: migrate
	docker compose -f $(COMPOSE_FILE) up -d postgres
	docker compose -f $(COMPOSE_FILE) up aster-scheduler

run-worker: migrate
	docker compose -f $(COMPOSE_FILE) up -d postgres
	docker compose -f $(COMPOSE_FILE) up aster-worker

# Development workflow with formatting
dev: format migrate test build run-demo

# Clean up everything
clean:
	rm -rf bin/
	docker compose -f $(COMPOSE_FILE) down -v

# Code Formatting (Go Standards)
format:
	@echo "Formatting Go code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		$(HOME)/go/bin/goimports -w .; \
	fi
	@echo "Code formatting completed"

# Check if code is properly
check-format:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not properly formatted:"; \
		gofmt -l .; \
		echo "Run 'make format' to fix formatting issues."; \
		exit 1; \
	fi
	@if [ -f $(HOME)/go/bin/goimports ]; then \
		if [ -n "$$($(HOME)/go/bin/goimports -l .)" ]; then \
			echo "The following files have import issues:"; \
			$(HOME)/go/bin/goimports -l .; \
			echo "Run 'make format' to fix import issues."; \
			exit 1; \
		fi \
	fi
	@echo "All files are properly formatted"

# Run go vet for code correctness
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "go vet passed"

# Tidy up go modules
tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "go mod tidy completed"

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Quick quality check (format + vet + tidy)
quality: format vet tidy
	@echo "Code quality check completed"

# Pre-commit hook
pre-commit: format quality test
	@echo "Pre-commit checks completed successfully"
