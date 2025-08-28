.PHONY: build test run-demo setup-db migrate clean dev full-demo

# Load environment variables from .env
include .env
export $(shell sed 's/=.*//' .env)

# Build all binaries
build:
	go build -o bin/aster-api ./cmd/aster-api
	go build -o bin/aster-scheduler ./cmd/aster-scheduler
	go build -o bin/aster-worker ./cmd/aster-worker

# Database setup
setup-db:
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

# Run migrations
migrate: setup-db
	@echo "Creating tables..."
	PGPASSWORD=$(POSTGRES_PASSWORD) psql -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f internal/db/migrations/001_jobs_table.sql
	PGPASSWORD=$(POSTGRES_PASSWORD) psql -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f internal/db/migrations/002_runs_table.sql

# Run all tests
test: migrate
	go test -v ./internal/config
	go test -v ./internal/db
	go test -v ./internal/db/store
	go test -v ./internal/scheduler
	go test -v ./internal/executor
	go test -v ./internal/worker

# Run the simple demo from earlier steps
run-demo: migrate build
	go run ./cmd/demo

# Full system build with all components
full-demo: migrate build
	chmod +x scripts/build.sh
	./scripts/build.sh

# Individual service commands
run-api: migrate build
	./bin/aster-api

run-scheduler: migrate build
	./bin/aster-scheduler

run-worker: migrate build
	./bin/aster-worker

# Development workflow
dev: migrate test build run-demo

# Clean up everything
clean:
	rm -rf bin/
	docker-compose down -v
