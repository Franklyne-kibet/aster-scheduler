# Aster Scheduler

[![Go Report Card](https://goreportcard.com/badge/github.com/Franklyne-kibet/aster-scheduler?refresh=1)](https://goreportcard.com/report/github.com/Franklyne-kibet/aster-scheduler)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-336791?logo=postgresql)](https://www.postgresql.org/)

A distributed, fault-tolerant job scheduler built with Go. Schedule and execute tasks using cron expressions with horizontal scaling, retry logic, and comprehensive monitoring.

## Table of Contents

- [Aster Scheduler](#aster-scheduler)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Architecture](#architecture)
  - [Quick Start](#quick-start)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Verify Installation](#verify-installation)
  - [Usage Examples](#usage-examples)
    - [Basic Job Creation](#basic-job-creation)
    - [Job with Environment Variables](#job-with-environment-variables)
    - [Monitoring Job Execution](#monitoring-job-execution)
  - [API Reference](#api-reference)
    - [Job Management](#job-management)
    - [Run Management](#run-management)
    - [System](#system)
  - [Configuration](#configuration)
  - [Documentation](#documentation)
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
- **Docker support** - Containerized deployment with docker-compose
- **High performance** - Built with Go for speed and concurrency

## Architecture

Aster Scheduler consists of three main components:

- **API Server** - REST API for job management and monitoring
- **Scheduler** - Monitors jobs and creates scheduled runs based on cron expressions
- **Worker** - Executes scheduled jobs and updates run status

All components communicate through a PostgreSQL database, providing fault tolerance and scalability.

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Make (optional, for using Makefile commands)

### Installation

```bash
# Clone the repository
git clone https://github.com/Franklyne-kibet/aster-scheduler.git
cd aster-scheduler

# Copy environment configuration
cp env.example .env

# Start all services
make full-demo
```

This will start all services and run a complete demonstration of the system.

### Verify Installation

```bash
# Check API health
curl http://localhost:8081/health

# Create a test job
curl -X POST http://localhost:8081/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test_job",
    "description": "Test job that runs every minute",
    "cron_expr": "* * * * *",
    "command": "echo",
    "args": ["Hello from Aster!"]
  }'

# List jobs
curl http://localhost:8081/api/v1/jobs

# List runs (after waiting ~65 seconds)
curl http://localhost:8081/api/v1/runs
```

## Usage Examples

### Basic Job Creation

Create a job that runs every hour:

```bash
curl -X POST http://localhost:8081/api/v1/jobs \
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

### Job with Environment Variables

```bash
curl -X POST http://localhost:8081/api/v1/jobs \
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

```bash
# List all jobs
curl http://localhost:8081/api/v1/jobs

# Get runs for a specific job
curl "http://localhost:8081/api/v1/runs?job_id=550e8400-e29b-41d4-a716-446655440000"

# Get recent runs
curl "http://localhost:8081/api/v1/runs?limit=10"
```

## API Reference

### Job Management

- **POST** `/api/v1/jobs` - Create a new job
- **GET** `/api/v1/jobs` - List all jobs
- **GET** `/api/v1/jobs/{id}` - Get specific job
- **PUT** `/api/v1/jobs/{id}` - Update job
- **DELETE** `/api/v1/jobs/{id}` - Delete job

### Run Management

- **GET** `/api/v1/runs` - List execution history
- **GET** `/api/v1/runs/{id}` - Get specific run details

### System

- **GET** `/health` - Health check endpoint

For detailed API documentation, see [docs/api-reference.md](docs/api-reference.md).

## Configuration

Aster Scheduler uses environment variables for configuration. Copy `env.example` to `.env` and modify as needed:

```bash
# Database configuration
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=aster

# API configuration
API_PORT=8080
API_EXTERNAL_PORT=8081
LOG_LEVEL=info

# Worker configuration
WORKER_POOL_SIZE=5
```

For detailed configuration options, see [docs/configuration.md](docs/configuration.md).

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Architecture Guide](docs/architecture.md)** - System design and component interactions
- **[Development Guide](docs/development.md)** - Setup, building, and contributing
- **[API Reference](docs/api-reference.md)** - Complete REST API documentation
- **[Deployment Guide](docs/deployment.md)** - Production deployment and operations
- **[Configuration Reference](docs/configuration.md)** - Environment variables and settings

## Contributing

We welcome contributions! Please see the [Development Guide](docs/development.md) for:

- Setting up a development environment
- Coding standards and quality requirements
- Testing guidelines
- Pull request process

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
