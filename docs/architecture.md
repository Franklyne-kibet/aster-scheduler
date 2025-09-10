# Architecture Guide

## System Overview

Aster Scheduler is a distributed job scheduling system with three main components:

- **API Server** - REST API for job management
- **Scheduler** - Monitors jobs and creates scheduled runs
- **Worker** - Executes scheduled jobs

All components communicate through PostgreSQL.

## Component Architecture

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Server    │    │   Scheduler     │    │     Worker      │
│   (aster-api)   │    │(aster-scheduler)│    │  (aster-worker) │
│                 │    │                 │    │                 │
│ • Job CRUD      │    │ • Cron parsing  │    │ • Job execution │
│ • Run queries   │    │ • Due job check │    │ • Status update │
│ • Health checks │    │ • Run creation  │    │ • Result storage│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                        ┌─────────────────┐
                        │   PostgreSQL    │
                        │   Database      │
                        │                 │
                        │ • Jobs table    │
                        │ • Runs table    │
                        └─────────────────┘
```

## Data Flow

### 1. Job Creation

```text
Client → API Server → Database
  │         │           │
  │ POST    │ INSERT    │ Store job
  │ /jobs   │ job       │ with next_run_at
  │         │           │
  │ ← 201   │ ← Success │
```

### 2. Job Scheduling

```text
    Scheduler → Database
    │         │
    │ Query   │ Return
    │ due     │ due jobs
    │ jobs    │
    │         │
    │ Create  │ Store
    │ runs    │ runs
    │         │
    │ Update  │ Store
    │ next    │ next_run_at
    │ run_at  │
```

### 3. Job Execution

```text
  Worker   → Database → Executor → OS
  │        │          │          │
  │ Query  │ Return   │ Execute  │ Run
  │ runs   │ runs     │ command  │ command
  │        │          │          │
  │ Update │ Store    │ Return   │
  │ status │ results  │ output   │
```

## Database Schema

### Jobs Table

```sql
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    cron_expr VARCHAR(255) NOT NULL,
    command VARCHAR(255) NOT NULL,
    args JSONB DEFAULT '[]',
    env JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    max_retries INTEGER DEFAULT 3,
    timeout INTERVAL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    next_run_at TIMESTAMP
);
```

### Runs Table

```sql
CREATE TABLE runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    attempt_num INTEGER DEFAULT 1,
    scheduled_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    output TEXT,
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Key Design Patterns

- **Database-Centric Coordination** - PostgreSQL as central state store
- **Event-Driven Processing** - Components react to database changes
- **Fault Tolerance** - Each component can restart independently
- **Horizontal Scaling** - Multiple worker instances
