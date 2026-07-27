package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
)

type workbenchRepository struct {
	db *sql.DB
}

func NewWorkbenchRepository(db *sql.DB) service.WorkbenchRepository {
	return &workbenchRepository{db: db}
}

func (r *workbenchRepository) ListModelBindings(ctx context.Context, userID int64) ([]service.WorkbenchModelBinding, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, api_key_id, model_id, display_name, provider, source, hidden, sort_order, created_at, updated_at
FROM workbench_model_bindings
WHERE user_id = $1
ORDER BY sort_order ASC, created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]service.WorkbenchModelBinding, 0)
	for rows.Next() {
		item, scanErr := scanWorkbenchModelBinding(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *workbenchRepository) GetModelBinding(ctx context.Context, userID int64, bindingID string) (*service.WorkbenchModelBinding, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, user_id, api_key_id, model_id, display_name, provider, source, hidden, sort_order, created_at, updated_at
FROM workbench_model_bindings
WHERE user_id = $1 AND id = $2
`, userID, bindingID)
	item, err := scanWorkbenchModelBinding(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrWorkbenchBindingNotFound
	}
	return item, err
}

func (r *workbenchRepository) EnsureModelBinding(ctx context.Context, binding service.WorkbenchModelBinding) (*service.WorkbenchModelBinding, error) {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO workbench_model_bindings (
    id, user_id, api_key_id, model_id, display_name, provider, source, hidden, sort_order
) VALUES ($1,$2,$3,$4,$5,$6,$7,FALSE,$8)
ON CONFLICT (user_id, api_key_id, model_id) DO NOTHING
`, binding.ID, binding.UserID, binding.APIKeyID, binding.ModelID, binding.DisplayName, binding.Provider, binding.Source, binding.SortOrder)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `
SELECT id, user_id, api_key_id, model_id, display_name, provider, source, hidden, sort_order, created_at, updated_at
FROM workbench_model_bindings
WHERE user_id = $1 AND api_key_id = $2 AND model_id = $3
`, binding.UserID, binding.APIKeyID, binding.ModelID)
	return scanWorkbenchModelBinding(row)
}

func (r *workbenchRepository) UpsertModelBinding(ctx context.Context, binding service.WorkbenchModelBinding) (*service.WorkbenchModelBinding, error) {
	row := r.db.QueryRowContext(ctx, `
INSERT INTO workbench_model_bindings (
    id, user_id, api_key_id, model_id, display_name, provider, source, hidden, sort_order
) VALUES ($1,$2,$3,$4,$5,$6,$7,FALSE,$8)
ON CONFLICT (user_id, api_key_id, model_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    provider = EXCLUDED.provider,
    source = EXCLUDED.source,
    hidden = FALSE,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW()
RETURNING id, user_id, api_key_id, model_id, display_name, provider, source, hidden, sort_order, created_at, updated_at
`, binding.ID, binding.UserID, binding.APIKeyID, binding.ModelID, binding.DisplayName, binding.Provider, binding.Source, binding.SortOrder)
	return scanWorkbenchModelBinding(row)
}

func (r *workbenchRepository) HideModelBinding(ctx context.Context, userID int64, bindingID string) error {
	result, err := r.db.ExecContext(ctx, `
UPDATE workbench_model_bindings
SET hidden = TRUE, updated_at = NOW()
WHERE user_id = $1 AND id = $2
`, userID, bindingID)
	if err != nil {
		return err
	}
	return workbenchRequireAffected(result, service.ErrWorkbenchBindingNotFound)
}

func (r *workbenchRepository) CreateConversation(ctx context.Context, conversation service.WorkbenchConversation) (*service.WorkbenchConversation, error) {
	row := r.db.QueryRowContext(ctx, `
INSERT INTO workbench_conversations (id, user_id, model_binding_id, title, reasoning_preset)
VALUES ($1,$2,$3,$4,$5)
RETURNING id, user_id, model_binding_id, title, reasoning_preset, created_at, updated_at
`, conversation.ID, conversation.UserID, conversation.ModelBindingID, conversation.Title, conversation.ReasoningPreset)
	return scanWorkbenchConversation(row)
}

func (r *workbenchRepository) ListConversations(ctx context.Context, userID int64, limit int) ([]service.WorkbenchConversation, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, model_binding_id, title, reasoning_preset, created_at, updated_at
FROM workbench_conversations
WHERE user_id = $1
ORDER BY updated_at DESC, created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]service.WorkbenchConversation, 0)
	for rows.Next() {
		item, scanErr := scanWorkbenchConversation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *workbenchRepository) GetConversation(ctx context.Context, userID int64, conversationID string) (*service.WorkbenchConversation, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, user_id, model_binding_id, title, reasoning_preset, created_at, updated_at
FROM workbench_conversations
WHERE user_id = $1 AND id = $2
`, userID, conversationID)
	item, err := scanWorkbenchConversation(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrWorkbenchConversationNotFound
	}
	return item, err
}

func (r *workbenchRepository) UpdateConversation(ctx context.Context, userID int64, conversationID string, patch service.WorkbenchConversationPatch) (*service.WorkbenchConversation, error) {
	row := r.db.QueryRowContext(ctx, `
UPDATE workbench_conversations
SET title = CASE WHEN $3 THEN $4 ELSE title END,
    model_binding_id = CASE WHEN $5 THEN $6 ELSE model_binding_id END,
    reasoning_preset = CASE WHEN $7 THEN $8 ELSE reasoning_preset END,
    updated_at = NOW()
WHERE user_id = $1 AND id = $2
RETURNING id, user_id, model_binding_id, title, reasoning_preset, created_at, updated_at
`, userID, conversationID,
		patch.Title != nil, stringValue(patch.Title),
		patch.ModelBindingID != nil, nullableStringValue(patch.ModelBindingID),
		patch.ReasoningPreset != nil, stringValue(patch.ReasoningPreset),
	)
	item, err := scanWorkbenchConversation(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrWorkbenchConversationNotFound
	}
	return item, err
}

func (r *workbenchRepository) DeleteConversation(ctx context.Context, userID int64, conversationID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM workbench_conversations WHERE user_id = $1 AND id = $2`, userID, conversationID)
	if err != nil {
		return err
	}
	return workbenchRequireAffected(result, service.ErrWorkbenchConversationNotFound)
}

func (r *workbenchRepository) ListMessages(ctx context.Context, userID int64, conversationID string) ([]service.WorkbenchMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT m.id, m.conversation_id, m.sequence, m.role, m.content, m.status, m.model_binding_id, m.reasoning_preset, m.reasoning_summary,
       m.error_message, m.request_id, m.attachments, m.created_at, m.updated_at
FROM workbench_messages m
JOIN workbench_conversations c ON c.id = m.conversation_id
WHERE c.user_id = $1 AND c.id = $2
ORDER BY m.sequence ASC
`, userID, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]service.WorkbenchMessage, 0)
	for rows.Next() {
		item, scanErr := scanWorkbenchMessage(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workbench_conversations WHERE user_id = $1 AND id = $2)`, userID, conversationID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, service.ErrWorkbenchConversationNotFound
		}
	}
	return items, nil
}

func (r *workbenchRepository) GetMessage(ctx context.Context, userID int64, messageID string) (*service.WorkbenchMessage, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT m.id, m.conversation_id, m.sequence, m.role, m.content, m.status, m.model_binding_id, m.reasoning_preset, m.reasoning_summary,
       m.error_message, m.request_id, m.attachments, m.created_at, m.updated_at
FROM workbench_messages m
JOIN workbench_conversations c ON c.id = m.conversation_id
WHERE c.user_id = $1 AND m.id = $2
`, userID, messageID)
	item, err := scanWorkbenchMessage(row)
	if err == sql.ErrNoRows {
		return nil, service.ErrWorkbenchMessageNotFound
	}
	return item, err
}

func (r *workbenchRepository) CreateTurn(ctx context.Context, userID int64, conversationID, content, title string, attachments []service.WorkbenchAttachment, modelBindingID *string, reasoningPreset string) (*service.WorkbenchTurn, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	conversationRow := tx.QueryRowContext(ctx, `
UPDATE workbench_conversations
SET title = $3,
    model_binding_id = $4,
    reasoning_preset = $5,
    updated_at = NOW()
WHERE user_id = $1 AND id = $2
RETURNING id, user_id, model_binding_id, title, reasoning_preset, created_at, updated_at
`, userID, conversationID, title, modelBindingID, reasoningPreset)
	conversation, err := scanWorkbenchConversation(conversationRow)
	if err == sql.ErrNoRows {
		return nil, service.ErrWorkbenchConversationNotFound
	}
	if err != nil {
		return nil, err
	}

	var sequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence), 0) FROM workbench_messages WHERE conversation_id = $1`, conversationID).Scan(&sequence); err != nil {
		return nil, err
	}
	userMessageID := uuid.NewString()
	assistantMessageID := uuid.NewString()
	attachmentsJSON, err := json.Marshal(attachments)
	if err != nil {
		return nil, err
	}
	userRow := tx.QueryRowContext(ctx, `
INSERT INTO workbench_messages (id, conversation_id, sequence, role, content, status, attachments)
VALUES ($1,$2,$3,'user',$4,$5,$6::jsonb)
RETURNING id, conversation_id, sequence, role, content, status, model_binding_id, reasoning_preset, reasoning_summary, error_message, request_id, attachments, created_at, updated_at
`, userMessageID, conversationID, sequence+1, content, service.WorkbenchMessageCompleted, string(attachmentsJSON))
	userMessage, err := scanWorkbenchMessage(userRow)
	if err != nil {
		return nil, err
	}
	assistantRow := tx.QueryRowContext(ctx, `
INSERT INTO workbench_messages (id, conversation_id, sequence, role, content, status, model_binding_id, reasoning_preset)
VALUES ($1,$2,$3,'assistant','',$4,$5,$6)
RETURNING id, conversation_id, sequence, role, content, status, model_binding_id, reasoning_preset, reasoning_summary, error_message, request_id, attachments, created_at, updated_at
`, assistantMessageID, conversationID, sequence+2, service.WorkbenchMessagePending, conversation.ModelBindingID, conversation.ReasoningPreset)
	assistantMessage, err := scanWorkbenchMessage(assistantRow)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &service.WorkbenchTurn{Conversation: conversation, UserMessage: userMessage, AssistantMessage: assistantMessage}, nil
}

func (r *workbenchRepository) ClaimAssistantMessage(ctx context.Context, userID int64, messageID, requestID string) error {
	result, err := r.db.ExecContext(ctx, `
UPDATE workbench_messages m
SET status = $3, request_id = $4, updated_at = NOW()
FROM workbench_conversations c
WHERE m.conversation_id = c.id
  AND c.user_id = $1
  AND m.id = $2
  AND m.role = 'assistant'
  AND m.status = $5
`, userID, messageID, service.WorkbenchMessageInProgress, requestID, service.WorkbenchMessagePending)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		message, getErr := r.GetMessage(ctx, userID, messageID)
		if getErr != nil {
			return getErr
		}
		if message.Status != service.WorkbenchMessagePending {
			return service.ErrWorkbenchGenerationStarted
		}
		return service.ErrWorkbenchMessageNotFound
	}
	return nil
}

func (r *workbenchRepository) FinishAssistantMessage(ctx context.Context, userID int64, messageID, status, content, reasoningSummary, errorMessage string) error {
	result, err := r.db.ExecContext(ctx, `
UPDATE workbench_messages m
SET status = $3,
    content = $4,
    reasoning_summary = $5,
    error_message = $6,
    updated_at = NOW()
FROM workbench_conversations c
WHERE m.conversation_id = c.id AND c.user_id = $1 AND m.id = $2 AND m.role = 'assistant'
  AND m.status <> $7
`, userID, messageID, status, content, reasoningSummary, errorMessage, service.WorkbenchMessageCanceled)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	message, err := r.GetMessage(ctx, userID, messageID)
	if err != nil {
		return err
	}
	if message.Status == service.WorkbenchMessageCanceled {
		return service.ErrWorkbenchGenerationCanceled
	}
	return service.ErrWorkbenchMessageNotFound
}

func (r *workbenchRepository) CancelAssistantMessage(ctx context.Context, userID int64, messageID string) error {
	result, err := r.db.ExecContext(ctx, `
UPDATE workbench_messages m
SET status = $3, error_message = 'Generation canceled', updated_at = NOW()
FROM workbench_conversations c
WHERE m.conversation_id = c.id
  AND c.user_id = $1
  AND m.id = $2
  AND m.role = 'assistant'
  AND m.status IN ($4, $5)
`, userID, messageID, service.WorkbenchMessageCanceled, service.WorkbenchMessagePending, service.WorkbenchMessageInProgress)
	if err != nil {
		return err
	}
	return workbenchRequireAffected(result, service.ErrWorkbenchMessageNotFound)
}

type workbenchScanner interface {
	Scan(dest ...any) error
}

func scanWorkbenchModelBinding(scanner workbenchScanner) (*service.WorkbenchModelBinding, error) {
	var item service.WorkbenchModelBinding
	err := scanner.Scan(&item.ID, &item.UserID, &item.APIKeyID, &item.ModelID, &item.DisplayName, &item.Provider,
		&item.Source, &item.Hidden, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func scanWorkbenchConversation(scanner workbenchScanner) (*service.WorkbenchConversation, error) {
	var item service.WorkbenchConversation
	var bindingID sql.NullString
	err := scanner.Scan(&item.ID, &item.UserID, &bindingID, &item.Title, &item.ReasoningPreset, &item.CreatedAt, &item.UpdatedAt)
	if bindingID.Valid {
		item.ModelBindingID = &bindingID.String
	}
	return &item, err
}

func scanWorkbenchMessage(scanner workbenchScanner) (*service.WorkbenchMessage, error) {
	var item service.WorkbenchMessage
	var attachmentsJSON []byte
	err := scanner.Scan(&item.ID, &item.ConversationID, &item.Sequence, &item.Role, &item.Content, &item.Status,
		&item.ModelBindingID, &item.ReasoningPreset, &item.ReasoningSummary, &item.ErrorMessage, &item.RequestID, &attachmentsJSON, &item.CreatedAt, &item.UpdatedAt)
	if err == nil && len(attachmentsJSON) > 0 {
		err = json.Unmarshal(attachmentsJSON, &item.Attachments)
	}
	return &item, err
}

func workbenchRequireAffected(result sql.Result, notFound error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound
	}
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nullableStringValue(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}

var _ service.WorkbenchRepository = (*workbenchRepository)(nil)
