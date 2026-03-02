CREATE TABLE IF NOT EXISTS aliases (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address          TEXT NOT NULL UNIQUE,
    label            TEXT,
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    emails_forwarded INTEGER NOT NULL DEFAULT 0,
    emails_blocked   INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_aliases_user_id ON aliases(user_id);
CREATE INDEX idx_aliases_address ON aliases(address);