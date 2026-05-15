CREATE TABLE IF NOT EXISTS reply_threads (
    reply_token         TEXT PRIMARY KEY,
    alias_id            UUID NOT NULL REFERENCES aliases(id) ON DELETE CASCADE,
    original_from       TEXT NOT NULL,
    original_message_id TEXT,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reply_threads_alias_id ON reply_threads(alias_id);
CREATE INDEX idx_reply_threads_expires_at ON reply_threads(expires_at);
