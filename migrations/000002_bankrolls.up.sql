CREATE TYPE currency_type AS ENUM ('BRL', 'USD', 'EUR', 'BTC');

CREATE TABLE IF NOT EXISTS bankrolls (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    currency currency_type NOT NULL,
    initial_balance NUMERIC(19, 4) NOT NULL,
    current_balance NUMERIC(19, 4) NOT NULL,
    start_date DATE NOT NULL,
    commission_percentage NUMERIC(5, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uq_bankroll_name_per_user UNIQUE (user_id, name),
    CONSTRAINT ck_initial_balance_nonnegative CHECK (initial_balance >= 0),
    CONSTRAINT ck_current_balance_nonnegative CHECK (current_balance >= 0),
    CONSTRAINT ck_commission_percentage_range CHECK (commission_percentage >= 0 AND commission_percentage <= 100)
);

CREATE INDEX IF NOT EXISTS idx_bankrolls_user_id ON bankrolls(user_id);
CREATE INDEX IF NOT EXISTS idx_bankrolls_deleted_at ON bankrolls(deleted_at);
