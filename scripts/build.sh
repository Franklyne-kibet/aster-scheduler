#!/bin/bash

# Complete Aster Demo Script
set -e

# Docker Compose settings
COMPOSE_FILE="deployments/docker-compose.yml"
DB_SERVICE="postgres"

# Load environment variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

echo "Aster Scheduler"
echo "======================="

# Start PostgreSQL
echo "Starting PostgreSQL..."
docker compose -f "$COMPOSE_FILE" up -d "$DB_SERVICE"
sleep 5

# Run migrations
echo "🗃️  Running database migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U $POSTGRES_USER -d $POSTGRES_DB -f internal/db/migrations/001_jobs_table.sql
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U $POSTGRES_USER -d $POSTGRES_DB -f internal/db/migrations/002_runs_table.sql

# Build binaries
echo "Building applications..."
go build -o bin/aster-api ./cmd/aster-api
go build -o bin/aster-scheduler ./cmd/aster-scheduler
go build -o bin/aster-worker ./cmd/aster-worker

# Start scheduler in background
echo "Starting scheduler..."
./bin/aster-scheduler &
SCHEDULER_PID=$!

# Start worker in background
echo "Starting worker..."
./bin/aster-worker &
WORKER_PID=$!

# Start API server in background
echo "Starting API server..."
./bin/aster-api &
API_PID=$!

# Wait for services to start
echo "Waiting for services to start..."
sleep 3

# Test API endpoints
echo ""
echo "Testing API endpoints..."

# Create a job
echo "Creating a job..."
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo_hello_world",
    "description": "A demo job that says hello",
    "cron_expr": "*/30 * * * *",
    "command": "echo",
    "args": ["Hello", "from", "Aster!"],
    "env": {"DEMO": "true"}
  }' | jq .

echo ""

# List jobs
echo "Listing jobs..."
curl -s http://localhost:8080/api/v1/jobs | jq .

echo ""

# Create another job that runs every minute
echo "Creating a frequent job..."
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo_date",
    "description": "Shows current date every minute",
    "cron_expr": "* * * * *",
    "command": "date",
    "args": []
  }' | jq .

echo ""
echo "Waiting 65 seconds for scheduler to pick up jobs..."
sleep 65

# List runs
echo "Listing runs..."
curl -s http://localhost:8080/api/v1/runs | jq .

echo ""
echo "✨ Demo complete!"
echo ""
echo "Try these commands:"
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8080/api/v1/jobs"
echo "  curl http://localhost:8080/api/v1/runs"

# Cleanup function
cleanup() {
  echo ""
  echo "Cleaning up..."
  kill $API_PID $SCHEDULER_PID $WORKER_PID 2>/dev/null || true
  docker compose -f "$COMPOSE_FILE" down
  echo "Cleanup complete"
}

# Wait for Ctrl+C
echo ""
echo "Press Ctrl+C to stop all services..."
trap cleanup EXIT
wait
