package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type imageGenerationRepository struct {
	db *sql.DB
}

func NewImageGenerationRepository(db *sql.DB) service.ImageGenerationRepository {
	return &imageGenerationRepository{db: db}
}

func (r *imageGenerationRepository) Create(ctx context.Context, params service.ImageGenerationCreateParams) (*service.ImageGeneration, error) {
	reqJSON := normalizeJSON(params.RequestJSON, []byte(`{}`))
	row := r.db.QueryRowContext(ctx, `
INSERT INTO image_generations (
    user_id, api_key_id, group_id, request_id, model, prompt, size, resolution_tier, aspect_ratio, quality, output_format, n, request_json
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb)
RETURNING id, user_id, api_key_id, group_id, request_id, upstream_request_id, model, prompt, size, resolution_tier, aspect_ratio, quality,
          output_format, n, status, storage_status, image_count, images, request_json, response_json,
          file_size_bytes, error_message, created_at, completed_at
`, params.UserID, params.APIKeyID, params.GroupID, params.RequestID, params.Model, params.Prompt, params.Size, params.ResolutionTier, params.AspectRatio, params.Quality, params.OutputFormat, params.N, string(reqJSON))
	return scanImageGeneration(row)
}

func (r *imageGenerationRepository) CreatePromptVersion(ctx context.Context, imageGenerationID int64, params service.ImageGenerationPromptVersion) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO image_generation_prompt_versions (
    image_generation_id, version, prompt, model, size, quality, output_format
) VALUES ($1,$2,$3,$4,$5,$6,$7)
`, imageGenerationID, params.Version, params.Prompt, params.Model, params.Size, params.Quality, params.OutputFmt)
	return err
}

func (r *imageGenerationRepository) UpdateResult(ctx context.Context, id int64, params service.ImageGenerationUpdateParams) (*service.ImageGeneration, error) {
	imagesJSON, err := json.Marshal(params.Images)
	if err != nil {
		return nil, fmt.Errorf("marshal image generation images: %w", err)
	}
	respJSON := normalizeJSON(params.ResponseJSON, []byte(`{}`))
	row := r.db.QueryRowContext(ctx, `
UPDATE image_generations
SET status = $2,
    storage_status = $3,
    images = $4::jsonb,
    response_json = $5::jsonb,
    file_size_bytes = $6,
    image_count = $7,
    error_message = $8,
    upstream_request_id = $9,
    completed_at = $10
WHERE id = $1
RETURNING id, user_id, api_key_id, group_id, request_id, upstream_request_id, model, prompt, size, resolution_tier, aspect_ratio, quality,
          output_format, n, status, storage_status, image_count, images, request_json, response_json,
          file_size_bytes, error_message, created_at, completed_at
`, id, params.Status, params.StorageStatus, string(imagesJSON), string(respJSON), params.FileSizeBytes, params.ImageCount, params.ErrorMessage, params.UpstreamRequestID, params.CompletedAt)
	return scanImageGeneration(row)
}

func (r *imageGenerationRepository) GetByID(ctx context.Context, id int64) (*service.ImageGeneration, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, user_id, api_key_id, group_id, request_id, upstream_request_id, model, prompt, size, resolution_tier, aspect_ratio, quality,
       output_format, n, status, storage_status, image_count, images, request_json, response_json,
       file_size_bytes, error_message, created_at, completed_at
FROM image_generations
WHERE id = $1
`, id)
	return scanImageGeneration(row)
}

func (r *imageGenerationRepository) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.ImageGeneration, *pagination.PaginationResult, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM image_generations WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, api_key_id, group_id, request_id, upstream_request_id, model, prompt, size, resolution_tier, aspect_ratio, quality,
       output_format, n, status, storage_status, image_count, images, request_json, response_json,
       file_size_bytes, error_message, created_at, completed_at
FROM image_generations
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
OFFSET $2 LIMIT $3
`, userID, params.Offset(), params.Limit())
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := make([]service.ImageGeneration, 0)
	for rows.Next() {
		item, err := scanImageGeneration(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *imageGenerationRepository) DeleteByUserID(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM image_generations WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrImageGenerationNotFound
	}
	return nil
}

func (r *imageGenerationRepository) ListPromptVersionsByUserID(ctx context.Context, userID, recordID int64) ([]service.ImageGenerationPromptVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT pv.version, pv.prompt, pv.model, pv.size, pv.quality, pv.output_format, pv.created_at
FROM image_generation_prompt_versions pv
JOIN image_generations ig ON ig.id = pv.image_generation_id
WHERE ig.user_id = $1 AND ig.id = $2
ORDER BY pv.version DESC, pv.id DESC
`, userID, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]service.ImageGenerationPromptVersion, 0)
	for rows.Next() {
		var item service.ImageGenerationPromptVersion
		if err := rows.Scan(&item.Version, &item.Prompt, &item.Model, &item.Size, &item.Quality, &item.OutputFmt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type imageGenerationScanner interface {
	Scan(dest ...any) error
}

func scanImageGeneration(scanner imageGenerationScanner) (*service.ImageGeneration, error) {
	var item service.ImageGeneration
	var apiKeyID sql.NullInt64
	var groupID sql.NullInt64
	var completedAt sql.NullTime
	var imagesRaw, requestRaw, responseRaw []byte
	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&apiKeyID,
		&groupID,
		&item.RequestID,
		&item.UpstreamRequestID,
		&item.Model,
		&item.Prompt,
		&item.Size,
		&item.ResolutionTier,
		&item.AspectRatio,
		&item.Quality,
		&item.OutputFormat,
		&item.N,
		&item.Status,
		&item.StorageStatus,
		&item.ImageCount,
		&imagesRaw,
		&requestRaw,
		&responseRaw,
		&item.FileSizeBytes,
		&item.ErrorMessage,
		&item.CreatedAt,
		&completedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrImageGenerationNotFound
		}
		return nil, err
	}
	if apiKeyID.Valid {
		item.APIKeyID = &apiKeyID.Int64
	}
	if groupID.Valid {
		item.GroupID = &groupID.Int64
	}
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	if len(imagesRaw) > 0 {
		_ = json.Unmarshal(imagesRaw, &item.Images)
	}
	item.RequestJSON = append([]byte(nil), requestRaw...)
	item.ResponseJSON = append([]byte(nil), responseRaw...)
	return &item, nil
}

func normalizeJSON(raw json.RawMessage, fallback []byte) []byte {
	if len(raw) == 0 || !json.Valid(raw) {
		return fallback
	}
	return raw
}
