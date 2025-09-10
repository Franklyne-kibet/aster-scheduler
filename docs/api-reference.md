# API Reference

## Base URL

- **Development**: `http://localhost:8081` (Docker)
- **Local**: `http://localhost:8080` (local)

## Authentication

No authentication required (implement for production).

## Job Management

### Create Job

```bash
POST /api/v1/jobs
```

**Request Body**:

```json
{
  "name": "string (required, unique)",
  "cron_expr": "string (required, cron expression)",
  "command": "string (required)",
  "args": ["string array (optional)"],
  "env": { "key": "value object (optional)" },
  "max_retries": "integer (optional, default: 3)",
  "timeout": "duration string (optional, e.g., '5m', '1h')"
}
```

**Response**: `201 Created`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "test_job",
  "cron_expr": "*/5 * * * *",
  "command": "echo",
  "args": ["Hello", "World"],
  "env": { "ENV_VAR": "value" },
  "status": "active",
  "max_retries": 3,
  "timeout": "5m",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "next_run_at": "2024-01-01T00:05:00Z"
}
```

### List Jobs

```bash
GET /api/v1/jobs
```

**Query Parameters**:

- `status` (optional) - Filter by status (`active`, `inactive`, `archived`)
- `limit` (optional) - Max results (default: 100)
- `offset` (optional) - Skip results (default: 0)

**Response**: `200 OK`

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "test_job",
    "cron_expr": "*/5 * * * *",
    "command": "echo",
    "args": ["Hello", "World"],
    "env": { "ENV_VAR": "value" },
    "status": "active",
    "max_retries": 3,
    "timeout": "5m",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "next_run_at": "2024-01-01T00:05:00Z"
  }
]
```

### Get Job

```bash
GET /api/v1/jobs/{id}
```

**Response**: `200 OK` (same as create job response)

### Update Job

```bash
PUT /api/v1/jobs/{id}
```

**Request Body**: Same as create job
**Response**: `200 OK` (updated job object)

### Delete Job

```bash
DELETE /api/v1/jobs/{id}
```

**Response**: `204 No Content`

## Run Management

### List Runs

```bash
GET /api/v1/runs
```

**Query Parameters**:

- `job_id` (optional) - Filter runs for specific job
- `status` (optional) - Filter by status (`scheduled`, `running`, `succeeded`, `failed`, `timed_out`, `cancelled`)
- `limit` (optional) - Max results (default: 100)
- `offset` (optional) - Skip results (default: 0)

**Response**: `200 OK`

```json
[
  {
    "id": "660f9511-f3ac-52e5-b827-557766551111",
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "succeeded",
    "attempt_num": 1,
    "scheduled_at": "2024-01-01T00:05:00Z",
    "started_at": "2024-01-01T00:05:01Z",
    "finished_at": "2024-01-01T00:05:02Z",
    "output": "Hello World\n",
    "error_msg": null,
    "created_at": "2024-01-01T00:05:00Z",
    "updated_at": "2024-01-01T00:05:02Z"
  }
]
```

### Get Run

```bash
GET /api/v1/runs/{id}
```

**Response**: `200 OK` (same as list runs response)

## System

### Health Check

```bash
GET /health
```

**Response**: `200 OK`

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Cron Expression Format

Standard 5-field cron expressions:

```text
┌───────────── minute (0 - 59)
│ ┌─────────── hour (0 - 23)
│ │ ┌───────── day of month (1 - 31)
│ │ │ ┌─────── month (1 - 12)
│ │ │ │ ┌───── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

**Examples**:

- `0 * * * *` - Every hour at minute 0
- `*/15 * * * *` - Every 15 minutes
- `0 9 * * 1-5` - Every weekday at 9:00 AM
- `0 0 1 * *` - First day of every month at midnight

## Error Responses

All errors return JSON with `error` field:

```json
{
  "error": "Error message description"
}
```

**Common HTTP Status Codes**:

- `400 Bad Request` - Invalid request data
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists (e.g., duplicate job name)
- `500 Internal Server Error` - Server error
