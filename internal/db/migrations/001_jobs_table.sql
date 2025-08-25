CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    cron_expr VARCHAR(100) NOT NULL,
    command TEXT NOT NULL,
    args JSONB DEFAULT '[]'::jsonb,
    env JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    max_retries INTEGER NOT NULL DEFAULT 3,
    timeout INTERVAL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    next_run_at TIMESTAMP WITH TIME ZONE
);

-- Index for common queries
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_next_run_at ON jobs(next_run_at) WHERE status = 'active';
