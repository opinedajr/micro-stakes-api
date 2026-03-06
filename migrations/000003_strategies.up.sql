CREATE TABLE IF NOT EXISTS strategies (
    id            BIGSERIAL    PRIMARY KEY,
    user_id       BIGINT       NOT NULL,
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    default_stake NUMERIC(15,2) NOT NULL,
    type          VARCHAR(4)   NOT NULL,
    active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT fk_strategy_user      FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT ck_strategy_type      CHECK (type IN ('Back', 'Lay')),
    CONSTRAINT ck_default_stake_pos  CHECK (default_stake > 0)
);

CREATE INDEX IF NOT EXISTS idx_strategies_user_id    ON strategies(user_id);
CREATE INDEX IF NOT EXISTS idx_strategies_created_at ON strategies(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_strategies_deleted_at ON strategies(deleted_at);
