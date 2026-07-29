CREATE TABLE IF NOT EXISTS workbench_attachments (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id VARCHAR(36) REFERENCES workbench_messages(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    object_key VARCHAR(768) NOT NULL UNIQUE,
    name VARCHAR(160) NOT NULL,
    mime_type VARCHAR(160) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'ready', 'attached')),
    etag VARCHAR(256) NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workbench_attachments_user_created
    ON workbench_attachments(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workbench_attachments_user_status_expiry
    ON workbench_attachments(user_id, status, expires_at);

CREATE INDEX IF NOT EXISTS idx_workbench_attachments_message
    ON workbench_attachments(message_id)
    WHERE message_id IS NOT NULL;
