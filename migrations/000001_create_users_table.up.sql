

BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(36)     PRIMARY KEY,
    name            VARCHAR(100)    NOT NULL,
    email           VARCHAR(255)    NOT NULL,
    password        VARCHAR(255)    NOT NULL,
    is_active       BOOLEAN         NOT NULL    DEFAULT TRUE,
    created_at      TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL    DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ     NULL
);

-- unique email constraint scoped to non-deleted users only
-- allows re-registration after account deletion
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active
    ON users (email)
    WHERE deleted_at IS NULL;

-- index for soft delete queries — most queries filter by deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_users_deleted_at
    ON users (deleted_at)
    WHERE deleted_at IS NULL;

-- index for active user lookups
CREATE INDEX IF NOT EXISTS idx_users_is_active
    ON users (is_active)
    WHERE deleted_at IS NULL;

COMMIT;