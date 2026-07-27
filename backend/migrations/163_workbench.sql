CREATE TABLE IF NOT EXISTS workbench_model_bindings (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    model_id VARCHAR(160) NOT NULL,
    display_name VARCHAR(160) NOT NULL DEFAULT '',
    provider VARCHAR(40) NOT NULL DEFAULT '',
    source VARCHAR(20) NOT NULL DEFAULT 'manual',
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, api_key_id, model_id)
);

CREATE TABLE IF NOT EXISTS workbench_conversations (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model_binding_id VARCHAR(36) REFERENCES workbench_model_bindings(id) ON DELETE SET NULL,
    title VARCHAR(160) NOT NULL DEFAULT '新对话',
    reasoning_preset VARCHAR(20) NOT NULL DEFAULT 'standard',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workbench_messages (
    id VARCHAR(36) PRIMARY KEY,
    conversation_id VARCHAR(36) NOT NULL REFERENCES workbench_conversations(id) ON DELETE CASCADE,
    sequence BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'completed',
    reasoning_summary TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    request_id VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (conversation_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_workbench_bindings_user_visible
    ON workbench_model_bindings(user_id, hidden, sort_order, created_at);

CREATE INDEX IF NOT EXISTS idx_workbench_conversations_user_updated
    ON workbench_conversations(user_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_workbench_messages_conversation_sequence
    ON workbench_messages(conversation_id, sequence);

CREATE INDEX IF NOT EXISTS idx_workbench_messages_request_id
    ON workbench_messages(request_id);
