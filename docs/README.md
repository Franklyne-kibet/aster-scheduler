# Aster Scheduler - Developer Documentation

This directory contains comprehensive documentation for developers working with the Aster Scheduler system.

## Documentation Structure

- **[Architecture Guide](architecture.md)** - System design, components, and data flow
- **[API Reference](api-reference.md)** - Complete REST API documentation
- **[Development Guide](development.md)** - Setup, building, and contributing
- **[Deployment Guide](deployment.md)** - Production deployment and operations
- **[Configuration Reference](configuration.md)** - Environment variables and settings

## Quick Navigation

### For New Developers

1. Start with [Architecture Guide](architecture.md) to understand the system
2. Follow [Development Guide](development.md) for setup
3. Reference [API Reference](api-reference.md) for integration

### For Contributors

1. Read [Development Guide](development.md) for coding standards
2. Review [Architecture Guide](architecture.md) for system understanding
3. Check [Configuration Reference](configuration.md) for environment setup

### For Operations Teams

1. Start with [Deployment Guide](deployment.md)
2. Reference [Configuration Reference](configuration.md)
3. Use [API Reference](api-reference.md) for monitoring

## System Overview

Aster Scheduler is a distributed job scheduling system built with Go. It consists of three main components:

- **API Server** - REST API for job management
- **Scheduler** - Monitors and schedules jobs based on cron expressions
- **Worker** - Executes scheduled jobs

All components communicate through a PostgreSQL database, providing fault tolerance and scalability.

## Key Features

- Cron-based job scheduling
- RESTful API for job management
- Horizontal worker scaling
- Fault tolerance and retry logic
- Complete execution history tracking
- Docker containerization support
- Environment variable configuration

## Getting Started

The fastest way to get started is with Docker:

```bash
# Clone the repository
git clone <repository-url>
cd aster-scheduler

# Copy environment configuration
cp env.example .env

# Run the complete system
make full-demo
```

This will start all services and run a demonstration of the system's capabilities.

## Architecture Highlights

- **Microservices Architecture** - Independent, scalable components
- **Database-Centric Coordination** - PostgreSQL as the central state store
- **Event-Driven Processing** - Scheduler triggers, workers respond
- **Fault-Tolerant Design** - Graceful error handling and recovery
- **Container-Ready** - Full Docker support with docker-compose

## Development Workflow

1. **Setup** - Use `make setup-db` to initialize the database
2. **Development** - Use `make dev` for the complete development workflow
3. **Testing** - Use `make test` to run the test suite
4. **Quality** - Use `make quality` for code quality checks
5. **Demo** - Use `make full-demo` to see the system in action

## Support

For questions about development, architecture, or contributing:

1. Check the relevant documentation in this directory
2. Review the code examples and tests
3. Examine the Makefile for available commands
4. Look at the configuration examples in `env.example`

## Next Steps

Choose the documentation that matches your needs:

- **Understanding the system** → [Architecture Guide](architecture.md)
- **Setting up development** → [Development Guide](development.md)
- **Integrating with the API** → [API Reference](api-reference.md)
- **Deploying to production** → [Deployment Guide](deployment.md)
