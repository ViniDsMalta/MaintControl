CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS machines (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),

    name VARCHAR(255) NOT NULL,

    type VARCHAR(50) NOT NULL,

    api_key TEXT UNIQUE NOT NULL,

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS production_lines (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),

    name VARCHAR(255) NOT NULL,

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS production_line_machines (
    id UUID PRIMARY KEY,

    production_line_id UUID NOT NULL REFERENCES production_lines(id) ON DELETE CASCADE,

    machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,

    position INTEGER NOT NULL,

    CONSTRAINT production_line_machine_unique UNIQUE (production_line_id, machine_id),
    CONSTRAINT production_line_position_unique UNIQUE (production_line_id, position)
);
CREATE TABLE IF NOT EXISTS telemetry (
    id BIGSERIAL PRIMARY KEY,

    machine_id UUID NOT NULL REFERENCES machines(id),

    temperature DOUBLE PRECISION,

    vibration DOUBLE PRECISION,

    rpm DOUBLE PRECISION,

    current DOUBLE PRECISION,

    pressure DOUBLE PRECISION,

    flow_rate DOUBLE PRECISION,

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS machine_status (
    machine_id UUID PRIMARY KEY REFERENCES machines(id),

    health_score DOUBLE PRECISION,

    risk_score DOUBLE PRECISION,

    status VARCHAR(50) not NULL DEFAULT 'Desconhecido',

    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_machine_time
ON telemetry(machine_id, created_at DESC);
