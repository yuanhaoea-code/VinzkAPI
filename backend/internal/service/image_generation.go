package service

import (
	"context"
	"encoding/json"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	ImageGenerationStatusProcessing = "processing"
	ImageGenerationStatusSuccess    = "success"
	ImageGenerationStatusFailed     = "failed"

	ImageStorageStatusPending = "pending"
	ImageStorageStatusSaved   = "saved"
	ImageStorageStatusFailed  = "failed"
)

var ErrImageGenerationNotFound = infraerrors.NotFound("IMAGE_GENERATION_NOT_FOUND", "image generation not found")
var ErrImageGenerationTierNotAllowed = infraerrors.Forbidden("IMAGE_TIER_NOT_ALLOWED", "selected API key does not allow the requested image resolution")

type ImageGenerationImage struct {
	Index     int    `json:"index"`
	Path      string `json:"path,omitempty"`
	URL       string `json:"url"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type ImageGeneration struct {
	ID                int64
	UserID            int64
	APIKeyID          *int64
	GroupID           *int64
	RequestID         string
	UpstreamRequestID string
	Model             string
	SourceModel       string
	Prompt            string
	PromptVersion     int64
	PromptHistory     []ImageGenerationPromptVersion
	Size              string
	ResolutionTier    string
	AspectRatio       string
	Quality           string
	OutputFormat      string
	N                 int
	Status            string
	StorageStatus     string
	ImageCount        int
	Images            []ImageGenerationImage
	RequestJSON       json.RawMessage
	ResponseJSON      json.RawMessage
	FileSizeBytes     int64
	ErrorMessage      string
	CreatedAt         time.Time
	CompletedAt       *time.Time
}

type ImageGenerationPromptVersion struct {
	Version   int64     `json:"version"`
	Prompt    string    `json:"prompt"`
	Model     string    `json:"model"`
	Size      string    `json:"size"`
	Quality   string    `json:"quality"`
	OutputFmt string    `json:"output_format"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateImageGenerationInput struct {
	UserID         int64
	APIKeyID       *int64
	APIKey         string
	Model          string
	Prompt         string
	SourceImage    string
	Size           string
	ResolutionTier string
	AspectRatio    string
	Quality        string
	OutputFormat   string
	N              int
}

type ImageGenerationCreateParams struct {
	UserID         int64
	APIKeyID       *int64
	GroupID        *int64
	RequestID      string
	Model          string
	Prompt         string
	SourceImage    string
	Size           string
	ResolutionTier string
	AspectRatio    string
	Quality        string
	OutputFormat   string
	N              int
	RequestJSON    json.RawMessage
}

type ImageGenerationUpdateParams struct {
	Status            string
	StorageStatus     string
	Images            []ImageGenerationImage
	ResponseJSON      json.RawMessage
	FileSizeBytes     int64
	ImageCount        int
	ErrorMessage      string
	UpstreamRequestID string
	CompletedAt       *time.Time
}

type ImageGenerationRepository interface {
	Create(ctx context.Context, params ImageGenerationCreateParams) (*ImageGeneration, error)
	CreatePromptVersion(ctx context.Context, imageGenerationID int64, params ImageGenerationPromptVersion) error
	UpdateResult(ctx context.Context, id int64, params ImageGenerationUpdateParams) (*ImageGeneration, error)
	GetByID(ctx context.Context, id int64) (*ImageGeneration, error)
	ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageGeneration, *pagination.PaginationResult, error)
	DeleteByUserID(ctx context.Context, userID, id int64) error
	ListPromptVersionsByUserID(ctx context.Context, userID, recordID int64) ([]ImageGenerationPromptVersion, error)
}
