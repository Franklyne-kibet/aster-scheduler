# Aster Scheduler

[![Build Status](https://github.com/Franklyne-kibet/aster-scheduler/workflows/CI/badge.svg)](https://github.com/Franklyne-kibet/aster-scheduler/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Franklyne-kibet/aster-scheduler)](https://goreportcard.com/report/github.com/Franklyne-kibet/aster-scheduler)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-336791?logo=postgresql)](https://www.postgresql.org/)

A distributed, fault-tolerant task scheduler with cron semantics, retries, job dependencies, and horizontally scalable worker pools.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [System Flow](#system-flow)
- [Flow Phases](#flow-phases)
- [Data Models](#data-models)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Development](#development)
- [Operational Concerns](#operational-concerns)
- [Security Considerations](#security-considerations)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Cron-based scheduling** - Standard cron expressions for flexible timing
- **Fault tolerance** - Automatic retries and error handling
- **Horizontal scaling** - Multiple workers can run concurrently
- **REST API** - Full CRUD operations for job management
- **Execution tracking** - Complete history of job runs
- **Timeout support** - Per-job execution timeouts
- **Environment variables** - Custom environment per job
- **High performance** - Built with Go for speed and concurrency

## Architecture

Aster Scheduler follows a microservices architecture with three main components that work together to provide a robust, scalable job scheduling system.

### Components

- **API Server** (`aster-api`) - REST API for job management
- **Scheduler** (`aster-scheduler`) - Monitors jobs and creates scheduled runs
- **Worker** (`aster-worker`) - Executes scheduled jobs

## System Flow

```text
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│                 │    │              │    │                 │
│   API CLIENT    │    │  API SERVER  │    │    DATABASE     │
│                 │    │   :8080      │    │   (PostgreSQL)  │
└─────────────────┘    └──────────────┘    └─────────────────┘
         │                       │                    │
         │ 1. POST /jobs         │                    │
         ├──────────────────────►│                    │
         │                       │ 2. INSERT job      │
         │                       ├───────────────────►│
         │                       │                    │
         │ 3. Job created (201)  │                    │
         | ◄─────────────────────┤                    │
         │                       │                    │
                                                      │
┌─────────────────┐    ┌──────────────┐               │
│                 │    │              │               │
│   SCHEDULER     │    │   JobStore   │               │
│  (every 30s)    │    │              │               │
└─────────────────┘    └──────────────┘               │
         │                       │                    │
         │ 4. Check due jobs     │                    │
         ├──────────────────────►│                    │
         │                       │ 5. SELECT jobs     │
         │                       │    WHERE next_run  │
         │                       │    <= NOW()        │
         │                       ├───────────────────►│
         │                       │                    │
         │ 6. Due jobs []        │ 7. Return jobs     │
          ◄──────────────────────┤ ◄──────────────────┤
         │                       │                    │
         │ 8. For each job:      │                    │
         │    - Create Run       │                    │
         │    - Update next_run  │                    │
         │                       │                    │
                                                      │
┌─────────────────┐    ┌──────────────┐               │
│                 │    │              │               │
│    RunStore     │    │    WORKER    │               │
│                 │    │  (every 5s)  │               │
└─────────────────┘    └──────────────┘               │
         │                       │                    │
         │ 9. INSERT run         │                    │
         ◄──────────────────────┤                     │
         │    (status=scheduled) │                    │
         ├───────────────────────┼───────────────────►│
         │                       │                    │
         │10. Poll scheduled     │                    │
         │    runs               │                    │
         ├──────────────────────►│                    │
         │                       │11. SELECT runs     │
         │                       │    WHERE status    │
         │                       │    = 'scheduled'   │
         │                       ├───────────────────►│
         │                       │                    │
         │12. Scheduled runs []  │13. Return runs     │
          ◄──────────────────────┤◄───────────────────┤
         │                       │                    │
         │14. For each run:      │                    │
         │    - Mark as running  │                    │
         │    - Execute job      │                    │
         │    - Mark finished    │                    │
         │                       │                    │
                                                      │
┌─────────────────┐    ┌──────────────┐               │
│                 │    │              │               │
│   EXECUTOR      │    │  OS COMMAND  │               │
│                 │    │              │               │
└─────────────────┘    └──────────────┘               │
         │                       │                    │
         │15. exec.Command(...)  │                    │
         ├──────────────────────►│                    │
         │                       │ 16. Run command    │
         │                       │     with args/env  │
         │                       │                    │
         │17. Output + Status    │                    │
          ◄──────────────────────┤                    │
         │                       │                    │
         │18. Update run status  │                    │
         │    and output         │                    │
         ├───────────────────────┼───────────────────►│
         │                       │                    │
```

## Flow Phases

### Phase 1: Job Creation

1. **Client** → **API**: `POST /api/v1/jobs` (with cron expression, command, etc.)
2. **API** → **Database**: Insert job into `jobs` table
3. **API** → **Client**: Return created job with ID

### Phase 2: Job Scheduling (Every 30 seconds)

1. **Scheduler** → **Database**: Query for due jobs (`next_run_at <= NOW()`)
2. **Scheduler** → **Database**: For each due job:
   - Insert record into `runs` table (status = "scheduled")
   - Update job's `next_run_at` based on cron expression

### Phase 3: Job Execution (Every 5 seconds)

1. **Worker** → **Database**: Query for scheduled runs (`status = 'scheduled'`)
2. **Worker** → **Database**: Mark run as "running"
3. **Worker** → **Executor**: Execute job command
4. **Executor** → **OS**: Run system command with args/env
5. **Executor** → **Worker**: Return output and exit status
6. **Worker** → **Database**: Update run status (succeeded/failed) and output

## Data Models

### Job Structure

```json
{
  "id": "uuid",
  "name": "unique_job_name",
  "description": "Job description",
  "cron_expr": "0 */10 * * *",
  "command": "echo",
  "args": ["Hello", "World"],
  "env": {"ENV_VAR": "value"},
  "status": "active",
  "max_retries": 3,
  "timeout": "5m",
  "next_run_at": "2024-01-01T10:00:00Z"
}
```

### Run Structure

```json
{
  "id": "uuid",
  "job_id": "uuid",
  "status": "succeeded",
  "attempt_num": 1,
  "scheduled_at": "2024-01-01T10:00:00Z",
  "started_at": "2024-01-01T10:00:01Z",
  "finished_at": "2024-01-01T10:00:05Z",
  "output": "Hello World\n",
  "error_msg": null
}
```

## Quick Start

### Prerequisites

Before setting up Aster Scheduler, ensure you have the following installed on your system:

- Go 1.21 or later
- PostgreSQL 13 or later
- Docker (optional, for containerized setup)
- Make (for using the provided Makefile)

### Installation and Setup

#### Option 1: Automated Setup

For the quickest way to get started, use the automated setup process:

```bash
# Clone the repository
git clone https://github.com/yourusername/aster-scheduler.git
cd aster-scheduler

# Copy and configure environment variables
cp .env.example .env
# Edit .env file with your database settings

# Set up database and run complete demo
make setup-db
make migrate
make full-demo
```

#### Option 2: Manual Setup

For more control over the setup process:

### Step 1: Database Setup

```bash
# Start PostgreSQL (if not already running)
sudo systemctl start postgresql

# Create database and user
createdb aster
createuser -P aster_user
```

### Step 2: Configuration

```bash
# Set environment variables
export DATABASE_URL="postgres://aster_user:password@localhost:5432/aster?sslmode=disable"
export API_PORT=8080
export LOG_LEVEL=info
```

### Step 3: Build and Run

```bash
# Build all components
make build

# Run database migrations
./bin/aster-migrate

# Start services (in separate terminals)
./bin/aster-api      # Terminal 1
./bin/aster-scheduler # Terminal 2
./bin/aster-worker   # Terminal 3
```

### Verification

Once all components are running, verify the system is working:

```bash
# Check API health
curl http://localhost:8080/health

# Create a test job
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test_job",
    "description": "Test job that runs every minute",
    "cron_expr": "* * * * *",
    "command": "echo",
    "args": ["Hello from Aster!"]
  }'
```

## API Reference

The REST API provides comprehensive job and run management capabilities. All endpoints return JSON responses and use standard HTTP status codes.

### Job Management Endpoints

#### Create Job

- **POST** `/api/v1/jobs`
- **Description**: Create a new scheduled job
- **Request Body**: Job object with name, cron_expr, command, and optional fields
- **Response**: 201 Created with job object, or 400 Bad Request for validation errors

#### List Jobs

- **GET** `/api/v1/jobs`
- **Description**: Retrieve all jobs with optional filtering
- **Query Parameters**:
  - `status` (optional): Filter by job status
  - `limit` (optional): Limit number of results (default: 100)
  - `offset` (optional): Skip number of results (default: 0)
- **Response**: 200 OK with array of job objects

#### Get Job

- **GET** `/api/v1/jobs/{id}`
- **Description**: Retrieve specific job by ID
- **Response**: 200 OK with job object, or 404 Not Found

#### Update Job

- **PUT** `/api/v1/jobs/{id}`
- **Description**: Update existing job
- **Request Body**: Partial or complete job object
- **Response**: 200 OK with updated job object, or 404 Not Found

### Delete Job

- **DELETE** `/api/v1/jobs/{id}`
- **Description**: Remove job and all associated runs
- **Response**: 204 No Content, or 404 Not Found

### Run Management Endpoints

#### List Runs

- **GET** `/api/v1/runs`
- **Description**: Retrieve execution history
- **Query Parameters**:
  - `job_id` (optional): Filter runs for specific job
  - `status` (optional): Filter by run status
  - `limit` (optional): Limit results (default: 100)
  - `offset` (optional): Skip results (default: 0)
- **Response**: 200 OK with array of run objects

#### Get Run

- **GET** `/api/v1/runs/{id}`
- **Description**: Retrieve specific run details
- **Response**: 200 OK with run object, or 404 Not Found

### System Endpoints

#### Health Check

- **GET** `/health`
- **Description**: System health status
- **Response**: 200 OK with health information

## Usage Examples

### Basic Job Creation

Create a simple job that runs every hour:

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hourly_cleanup",
    "description": "Clean temporary files every hour",
    "cron_expr": "0 * * * *",
    "command": "find",
    "args": ["/tmp", "-name", "*.tmp", "-delete"],
    "timeout": "5m"
  }'
```

### Advanced Job with Environment

Create a job with custom environment variables and retry logic:

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api_healthcheck",
    "description": "Check API endpoint health every 5 minutes",
    "cron_expr": "*/5 * * * *",
    "command": "curl",
    "args": ["-f", "-s", "-o", "/dev/null", "$API_ENDPOINT"],
    "env": {
      "API_ENDPOINT": "https://api.example.com/health",
      "HTTP_TIMEOUT": "30"
    },
    "max_retries": 2,
    "timeout": "1m"
  }'
```

### Monitoring Job Execution

Check job execution history:

```bash
# List all recent runs
curl "http://localhost:8080/api/v1/runs?limit=10"

# Check runs for specific job
curl "http://localhost:8080/api/v1/runs?job_id=550e8400-e29b-41d4-a716-446655440000"

# Get detailed run information
curl "http://localhost:8080/api/v1/runs/660f9511-f3ac-52e5-b827-557766551111"
```

## Configuration

Aster Scheduler uses environment variables for configuration, allowing easy deployment across different environments.

### Database Configuration

```bash
# Primary database connection
DATABASE_URL=postgres://username:password@hostname:5432/database?sslmode=disable

# Alternative individual settings
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=aster_user
POSTGRES_PASSWORD=secure_password
POSTGRES_DB=aster
POSTGRES_SSL_MODE=disable
```

### Service Configuration

```bash
# API Server settings
API_PORT=8080
API_HOST=0.0.0.0

# Scheduler settings
SCHEDULER_INTERVAL=30s

# Worker settings
WORKER_POLL_INTERVAL=5s
WORKER_CONCURRENT_JOBS=5

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Timeout and Retry Settings

```bash
# Default job timeout (can be overridden per job)
DEFAULT_JOB_TIMEOUT=10m

# Default retry attempts
DEFAULT_MAX_RETRIES=3

# Database connection timeout
DB_TIMEOUT=30s
```

## Development

### Development Environment Setup

Setting up a development environment for contributing to Aster Scheduler:

```bash
# Clone and setup
git clone https://github.com/yourusername/aster-scheduler.git
cd aster-scheduler

# Install development dependencies
go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Setup development database
make setup-dev-db

# Run in development mode
make dev
```

### Available Make Targets

The Makefile provides convenient commands for common development tasks:

- `make build` - Build all service binaries
- `make test` - Run the complete test suite
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run code linting and formatting checks
- `make migrate` - Apply database migrations
- `make setup-db` - Initialize development database
- `make run-demo` - Run basic functionality demo
- `make full-demo` - Run comprehensive system demo
- `make clean` - Remove build artifacts and containers
- `make dev` - Complete development workflow (build, test, lint)

### Testing

The project includes comprehensive tests for all components:

```bash
# Run all tests
make test

# Run tests for specific package
go test -v ./internal/scheduler
go test -v ./internal/worker
go test -v ./internal/api

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Operational Concerns

### Context and Cancellation

Aster Scheduler implements comprehensive context handling for robust operation. All long-running operations respect context cancellation, enabling graceful shutdowns and preventing resource leaks. Job executions can be cancelled through timeouts, and the system responds appropriately to shutdown signals.

### Fault Tolerance

The system is designed with fault tolerance as a primary concern. Each component can restart independently without affecting the others, since all state is persisted in PostgreSQL. Failed jobs are automatically retried based on their configuration, and one job's failure doesn't impact other scheduled work.

### Performance and Scaling

For production deployments, consider these scaling strategies:

**Horizontal Worker Scaling**: Deploy multiple worker instances to handle increased job volume. Workers coordinate through the database to prevent duplicate execution.

**Scheduler Optimization**: Adjust the scheduler interval based on your timing requirements. More frequent checks provide better precision but increase database load.

**Database Optimization**: Use connection pooling and appropriate PostgreSQL configuration for your workload. Consider read replicas for read-heavy monitoring operations.

**Resource Management**: Set appropriate timeouts and resource limits for jobs to prevent resource exhaustion. Monitor system resources and scale accordingly.

### Monitoring and Observability

Implement proper monitoring for production deployments:

- Monitor component health endpoints
- Track job success/failure rates
- Alert on stuck or repeatedly failing jobs
- Monitor database performance and connection usage
- Set up log aggregation for troubleshooting

## Security Considerations

When deploying Aster Scheduler in production environments, consider these security aspects:

- Secure database connections with SSL and proper authentication
- Limit command execution capabilities through system-level controls
- Validate job commands and arguments to prevent injection attacks
- Use network segmentation to isolate scheduler components
- Implement proper access controls for the REST API
- Regularly update dependencies and base images

## Contributing

We welcome contributions to Aster Scheduler. To contribute effectively:

1. **Fork the repository** and create a feature branch from main
2. **Follow the coding standards** established in the project
3. **Add comprehensive tests** for any new functionality
4. **Update documentation** for user-facing changes
5. **Run the full test suite** and ensure all checks pass
6. **Submit a pull request** with a clear description of changes

Please review the existing code and tests to understand the patterns and conventions used in the project.

## License

This project is licensed under the Apache License. See the LICENSE file for complete license terms and conditions.
