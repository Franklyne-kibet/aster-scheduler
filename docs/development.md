# Development Guide

## Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **PostgreSQL 13+** - [Download](https://www.postgresql.org/download/)
- **Docker & Docker Compose** - [Download](https://www.docker.com/get-started)
- **Make** - Usually pre-installed on Unix systems

## Quick Start

```bash
# Clone and setup
git clone https://github.com/Franklyne-kibet/aster-scheduler.git
cd aster-scheduler
cp env.example .env

# Start development environment
make setup-db
make migrate
make full-demo
```

## Project Structure

```text
aster-scheduler/
├── cmd/                    # Application entry points
│   ├── aster-api/         # API server
│   ├── aster-scheduler/   # Scheduler
│   └── aster-worker/      # Worker
├── internal/              # Private application code
│   ├── api/               # HTTP API
│   ├── config/            # Configuration
│   ├── db/                # Database layer
│   ├── executor/          # Job execution
│   ├── scheduler/         # Scheduling logic
│   ├── types/             # Data structures
│   └── worker/            # Worker implementation
├── docs/                  # Documentation
├── scripts/               # Build scripts
└── docker-compose.yml     # Docker services
```

## Available Commands

```bash
# Build
make build                 # Build all binaries
make build-api             # Build API server
make build-scheduler       # Build scheduler
make build-worker          # Build worker

# Database
make setup-db              # Start PostgreSQL container
make migrate               # Run database migrations
make clean-db              # Clean database

# Run services
make run-api               # Run API server
make run-scheduler         # Run scheduler
make run-worker            # Run worker
make full-demo             # Run complete demo

# Development
make test                  # Run tests
make lint                  # Run linter
make fmt                   # Format code
make clean                 # Clean build artifacts
```

## Development Workflow

1. **Start database**:

   ```bash
   make setup-db
   make migrate
   ```

2. **Run services individually**:

   ```bash
   make run-api
   make run-scheduler
   make run-worker
   ```

3. **Or run everything**:

   ```bash
   make full-demo
   ```

## Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test
```

## Testing

```bash
# Run all tests
make test

# Run specific test
go test ./internal/scheduler/...

# Run with coverage
go test -cover ./...
```

## Building

```bash
# Build all services
make build

# Build specific service
make build-api

# Build for production
make build-all
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linter
5. Submit a pull request

## Code Standards

- Use `go fmt` for formatting
- Follow Go naming conventions
- Add tests for new functionality
- Update documentation for API changes
