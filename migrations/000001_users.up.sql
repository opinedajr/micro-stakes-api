CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    identity_id VARCHAR(255) NOT NULL UNIQUE,
    identity_adapter VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_identity_id ON users(identity_id);
CREATE INDEX IF NOT EXISTS idx_users_identity_adapter ON users(identity_adapter);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
