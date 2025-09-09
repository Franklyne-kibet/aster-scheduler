#!/bin/bash

# Complete Aster Demo Script
set -e

# Docker Compose settings
COMPOSE_FILE="docker-compose.yml"
DB_SERVICE="postgres"

# Load environment variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

echo "Aster Scheduler"
echo "======================="

# Start all services with Docker Compose
echo "Starting all services with Docker Compose..."
docker compose -f "$COMPOSE_FILE" up -d

# Wait for services to start
echo "Waiting for services to start..."
sleep 10

# Test API endpoints
echo ""
echo "Testing API endpoints..."

# Create a job
echo "Creating a job..."
curl -X POST http://localhost:8081/api/v1/jobs \
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
curl -s http://localhost:8081/api/v1/jobs | jq .

echo ""

# Create another job that runs every minute
echo "Creating a frequent job..."
curl -X POST http://localhost:8081/api/v1/jobs \
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
curl -s http://localhost:8081/api/v1/runs | jq .

echo ""
echo "✨ Demo complete!"
echo ""
echo "Try these commands:"
echo "  curl http://localhost:8081/health"
echo "  curl http://localhost:8081/api/v1/jobs"
echo "  curl http://localhost:8081/api/v1/runs"

# Cleanup function
cleanup() {
  echo ""
  echo "Cleaning up..."
  docker compose -f "$COMPOSE_FILE" down
  echo "Cleanup complete"
}

# Wait for Ctrl+C
echo ""
echo "Press Ctrl+C to stop all services..."
trap cleanup EXIT
wait
