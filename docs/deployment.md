# Deployment Guide

## Quick Start

```bash
# Clone and setup
git clone https://github.com/Franklyne-kibet/aster-scheduler.git
cd aster-scheduler
cp env.example .env

# Deploy all services
docker compose up -d

# Check status
docker compose ps
```

## Development Deployment

```bash
# Start with demo
make full-demo

# Or start individual services
make run-api
make run-scheduler
make run-worker
```

## Production Deployment

```bash
# Production environment
cp env.example .env.production
# Edit .env.production with production values

# Deploy with scaling
docker compose -f docker-compose.yml --env-file .env.production up -d --scale aster-worker=3
```

## Database Setup

Migrations run automatically via Makefile:

```bash
make migrate
```

## Health Checks

```bash
# API health
curl http://localhost:8081/health

# Database health
docker compose exec postgres pg_isready -U postgres
```

## Logs

```bash
# All logs
docker compose logs -f

# Specific service
docker compose logs -f aster-api
```

## Scaling

```bash
# Scale workers
docker compose up -d --scale aster-worker=3

# Scale API (if needed)
docker compose up -d --scale aster-api=2
```

## Troubleshooting

**Port conflicts**:

```bash
lsof -i :8081
lsof -i :5432
```

**Database issues**:

```bash
docker compose logs postgres
docker compose exec postgres psql -U postgres -d aster -c "SELECT 1;"
```

**Service not starting**:

```bash
docker compose logs aster-api
docker compose exec aster-api /bin/sh
```

## Maintenance

**Backup database**:

```bash
docker compose exec postgres pg_dump -U postgres aster > backup.sql
```

**Update services**:

```bash
docker compose down
docker compose build
docker compose up -d
make migrate
```
