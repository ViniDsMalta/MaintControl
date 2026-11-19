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

    type VARCHAR(50) NOT NULL CHECK (type IN ('motor', 'bomba', 'compressor')),

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

    position INTEGER NOT NULL CHECK (position > 0),

    CONSTRAINT production_line_machine_unique UNIQUE (production_line_id, machine_id),
    CONSTRAINT production_line_position_unique UNIQUE (production_line_id, position)
);
CREATE TABLE IF NOT EXISTS telemetry (
    id BIGSERIAL PRIMARY KEY,

    machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,

    temperature DOUBLE PRECISION CHECK (temperature BETWEEN 30 AND 115),

    vibration DOUBLE PRECISION CHECK (vibration BETWEEN 0.1 AND 8),

    rpm DOUBLE PRECISION CHECK (rpm BETWEEN 900 AND 2200),

    current DOUBLE PRECISION,

    pressure DOUBLE PRECISION CHECK (pressure BETWEEN 0.3 AND 12),

    flow_rate DOUBLE PRECISION CHECK (flow_rate BETWEEN 10 AND 130),

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS machine_status (
    machine_id UUID PRIMARY KEY REFERENCES machines(id) ON DELETE CASCADE,

    health_score DOUBLE PRECISION CHECK (health_score BETWEEN 0 AND 100),

    risk_score DOUBLE PRECISION CHECK (risk_score BETWEEN 0 AND 100),

    status VARCHAR(50) NOT NULL DEFAULT 'AGUARDANDO_DADOS'
        CHECK (status IN ('AGUARDANDO_DADOS', 'NORMAL', 'DEGRADATION', 'FAILURE')),

    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_machine_time
ON telemetry(machine_id, created_at DESC);
