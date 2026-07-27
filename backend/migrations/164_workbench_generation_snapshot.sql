ALTER TABLE workbench_messages
    ADD COLUMN IF NOT EXISTS model_binding_id VARCHAR(36) REFERENCES workbench_model_bindings(id) ON DELETE SET NULL;

ALTER TABLE workbench_messages
    ADD COLUMN IF NOT EXISTS reasoning_preset VARCHAR(20) NOT NULL DEFAULT 'standard';

UPDATE workbench_messages AS m
SET model_binding_id = c.model_binding_id,
    reasoning_preset = c.reasoning_preset
FROM workbench_conversations AS c
WHERE m.conversation_id = c.id
  AND m.role = 'assistant'
  AND m.model_binding_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_workbench_messages_model_binding
    ON workbench_messages(model_binding_id);
