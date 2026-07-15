CREATE TABLE IF NOT EXISTS image_generations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    request_id VARCHAR(128) NOT NULL DEFAULT '',
    upstream_request_id VARCHAR(128) NOT NULL DEFAULT '',
    model VARCHAR(128) NOT NULL DEFAULT '',
    prompt TEXT NOT NULL DEFAULT '',
    size VARCHAR(64) NOT NULL DEFAULT '',
    quality VARCHAR(64) NOT NULL DEFAULT '',
    output_format VARCHAR(32) NOT NULL DEFAULT '',
    n INT NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL DEFAULT 'processing',
    storage_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    image_count INT NOT NULL DEFAULT 0,
    images JSONB NOT NULL DEFAULT '[]'::jsonb,
    request_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    file_size_bytes BIGINT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS image_generation_prompt_versions (
    id BIGSERIAL PRIMARY KEY,
    image_generation_id BIGINT NOT NULL REFERENCES image_generations(id) ON DELETE CASCADE,
    version BIGINT NOT NULL DEFAULT 1,
    prompt TEXT NOT NULL DEFAULT '',
    model VARCHAR(128) NOT NULL DEFAULT '',
    size VARCHAR(64) NOT NULL DEFAULT '',
    quality VARCHAR(64) NOT NULL DEFAULT '',
    output_format VARCHAR(32) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_generations_user_created_at
    ON image_generations(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_image_generations_user_status
    ON image_generations(user_id, status);

CREATE INDEX IF NOT EXISTS idx_image_generations_request_id
    ON image_generations(request_id);

CREATE INDEX IF NOT EXISTS idx_image_generation_prompt_versions_generation_id_version
    ON image_generation_prompt_versions(image_generation_id, version DESC);
