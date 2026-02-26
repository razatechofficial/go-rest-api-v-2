BEGIN;

CREATE TABLE IF NOT EXISTS orders (
    id                  VARCHAR(36)     PRIMARY KEY,
    user_id             VARCHAR(36)     NOT NULL REFERENCES users(id),
    status              VARCHAR(32)     NOT NULL DEFAULT 'pending',
    total_amount_cents  BIGINT          NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ     NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id
    ON orders (user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_deleted_at
    ON orders (deleted_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_created_at
    ON orders (created_at DESC)
    WHERE deleted_at IS NULL;

COMMIT;
