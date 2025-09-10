# Configuration Reference

## Environment Variables

All configuration is done through environment variables. Copy `env.example` to `.env` and modify as needed.

## Database Configuration

| Variable                 | Default    | Description                                                          |
| ------------------------ | ---------- | -------------------------------------------------------------------- |
| `POSTGRES_HOST`          | `postgres` | Database hostname (use `postgres` for Docker, `localhost` for local) |
| `POSTGRES_PORT`          | `5432`     | Database port                                                        |
| `POSTGRES_USER`          | `postgres` | Database username                                                    |
| `POSTGRES_PASSWORD`      | `postgres` | Database password                                                    |
| `POSTGRES_DB`            | `aster`    | Database name                                                        |
| `POSTGRES_EXTERNAL_PORT` | `5432`     | External port for database access                                    |

## API Server Configuration

| Variable            | Default                                       | Description                        |
| ------------------- | --------------------------------------------- | ---------------------------------- |
| `API_PORT`          | `8080`                                        | Internal API port                  |
| `API_EXTERNAL_PORT` | `8081`                                        | External API port (Docker mapping) |
| `ALLOWED_ORIGINS`   | `http://localhost:3000,http://127.0.0.1:3000` | CORS allowed origins               |
| `READ_TIMEOUT`      | `15s`                                         | HTTP read timeout                  |
| `WRITE_TIMEOUT`     | `15s`                                         | HTTP write timeout                 |
| `IDLE_TIMEOUT`      | `60s`                                         | HTTP idle timeout                  |

## Worker Configuration

| Variable           | Default | Description                        |
| ------------------ | ------- | ---------------------------------- |
| `WORKER_POOL_SIZE` | `5`     | Maximum concurrent jobs per worker |

## Scheduler Configuration

| Variable              | Default | Description                       |
| --------------------- | ------- | --------------------------------- |
| `LEADER_ELECTION_TTL` | `30s`   | Scheduler leader election timeout |

## Logging Configuration

| Variable    | Default | Description                                  |
| ----------- | ------- | -------------------------------------------- |
| `LOG_LEVEL` | `info`  | Log level (`debug`, `info`, `warn`, `error`) |

## Example Configuration

```bash
# Database
POSTGRES_USER=aster_user
POSTGRES_PASSWORD=secure_password
POSTGRES_DB=aster_scheduler
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_EXTERNAL_PORT=5432

# API
API_PORT=8080
API_EXTERNAL_PORT=8081
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

# Worker
WORKER_POOL_SIZE=10

# Logging
LOG_LEVEL=debug
```

## Configuration Management

Configuration is managed through:

- **Environment variables** in `.env` files
- **Docker Compose** for containerized deployments
- **Makefile** for development and build automation
