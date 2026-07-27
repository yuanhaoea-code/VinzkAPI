package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageGenerationHandler struct {
	service *service.ImageGenerationService
}

func NewImageGenerationHandler(service *service.ImageGenerationService) *ImageGenerationHandler {
	return &ImageGenerationHandler{service: service}
}

func (h *ImageGenerationHandler) Pricing(c *gin.Context) {
	response.Success(c, gin.H{
		"currency": "USD",
		"tiers":    service.ImageGenerationFixedPrices(),
	})
}

func (h *ImageGenerationHandler) Capabilities(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	capabilities, err := h.service.ListImageGenerationCapabilities(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, capabilities)
}

type createImageGenerationRequest struct {
	APIKeyID       *int64 `json:"api_key_id"`
	APIKey         string `json:"api_key"`
	Model          string `json:"model"`
	Prompt         string `json:"prompt" binding:"required"`
	SourceImage    string `json:"source_image"`
	Size           string `json:"size"`
	ResolutionTier string `json:"resolution_tier"`
	AspectRatio    string `json:"aspect_ratio"`
	Quality        string `json:"quality"`
	OutputFormat   string `json:"output_format"`
	N              int    `json:"n"`
}

type imageGenerationResponse struct {
	ID                int64                          `json:"id"`
	APIKeyID          *int64                         `json:"api_key_id,omitempty"`
	GroupID           *int64                         `json:"group_id,omitempty"`
	RequestID         string                         `json:"request_id"`
	UpstreamRequestID string                         `json:"upstream_request_id,omitempty"`
	Model             string                         `json:"model"`
	Prompt            string                         `json:"prompt"`
	Size              string                         `json:"size"`
	ResolutionTier    string                         `json:"resolution_tier"`
	AspectRatio       string                         `json:"aspect_ratio"`
	Quality           string                         `json:"quality"`
	OutputFormat      string                         `json:"output_format"`
	N                 int                            `json:"n"`
	Status            string                         `json:"status"`
	StorageStatus     string                         `json:"storage_status"`
	ImageCount        int                            `json:"image_count"`
	Images            []service.ImageGenerationImage `json:"images"`
	FileSizeBytes     int64                          `json:"file_size_bytes"`
	ErrorMessage      string                         `json:"error_message,omitempty"`
	CreatedAt         string                         `json:"created_at"`
	CompletedAt       *string                        `json:"completed_at,omitempty"`
}

type imageGenerationPromptVersionResponse struct {
	Version      int64  `json:"version"`
	Prompt       string `json:"prompt"`
	Model        string `json:"model"`
	Size         string `json:"size"`
	Quality      string `json:"quality"`
	OutputFormat string `json:"output_format"`
	CreatedAt    string `json:"created_at"`
}

func (h *ImageGenerationHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createImageGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	record, err := h.service.Generate(c.Request.Context(), service.CreateImageGenerationInput{
		UserID:         subject.UserID,
		APIKeyID:       req.APIKeyID,
		APIKey:         req.APIKey,
		Model:          req.Model,
		Prompt:         req.Prompt,
		SourceImage:    req.SourceImage,
		Size:           req.Size,
		ResolutionTier: req.ResolutionTier,
		AspectRatio:    req.AspectRatio,
		Quality:        req.Quality,
		OutputFormat:   req.OutputFormat,
		N:              req.N,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, mapImageGeneration(record))
}

func (h *ImageGenerationHandler) CreateQueued(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createImageGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	record, err := h.service.GenerateQueued(c.Request.Context(), service.CreateImageGenerationInput{
		UserID:         subject.UserID,
		APIKeyID:       req.APIKeyID,
		APIKey:         req.APIKey,
		Model:          req.Model,
		Prompt:         req.Prompt,
		SourceImage:    req.SourceImage,
		Size:           req.Size,
		ResolutionTier: req.ResolutionTier,
		AspectRatio:    req.AspectRatio,
		Quality:        req.Quality,
		OutputFormat:   req.OutputFormat,
		N:              req.N,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, mapImageGeneration(record))
}

func (h *ImageGenerationHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.List(c.Request.Context(), subject.UserID, pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "created_at",
		SortOrder: "desc",
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]imageGenerationResponse, 0, len(items))
	for i := range items {
		out = append(out, mapImageGeneration(&items[i]))
	}
	response.Paginated(c, out, result.Total, result.Page, result.PageSize)
}

func (h *ImageGenerationHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid image generation ID")
		return
	}
	record, err := h.service.GetOwned(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, mapImageGeneration(record))
}

func (h *ImageGenerationHandler) Delete(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid image generation ID")
		return
	}
	if err := h.service.Delete(c.Request.Context(), subject.UserID, id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ImageGenerationHandler) PromptVersions(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid image generation ID")
		return
	}
	versions, err := h.service.GetPromptVersions(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]imageGenerationPromptVersionResponse, 0, len(versions))
	for _, item := range versions {
		out = append(out, imageGenerationPromptVersionResponse{
			Version:      item.Version,
			Prompt:       item.Prompt,
			Model:        item.Model,
			Size:         item.Size,
			Quality:      item.Quality,
			OutputFormat: item.OutputFmt,
			CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	response.Success(c, gin.H{"items": out})
}

func (h *ImageGenerationHandler) Preview(c *gin.Context) {
	h.serveImage(c, false)
}

func (h *ImageGenerationHandler) Download(c *gin.Context) {
	h.serveImage(c, true)
}

func (h *ImageGenerationHandler) serveImage(c *gin.Context, download bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid image generation ID")
		return
	}
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil {
		response.BadRequest(c, "Invalid image index")
		return
	}
	file, img, err := h.service.OpenImage(c.Request.Context(), subject.UserID, id, index)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	defer file.Close()

	if img.MimeType != "" {
		c.Header("Content-Type", img.MimeType)
	}
	if download {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="vinzkapi-image-%d-%d.%s"`, id, index, imageDownloadExt(img.MimeType)))
	}
	c.DataFromReader(http.StatusOK, -1, img.MimeType, file, nil)
}

func mapImageGeneration(record *service.ImageGeneration) imageGenerationResponse {
	var completed *string
	if record.CompletedAt != nil {
		v := record.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		completed = &v
	}
	return imageGenerationResponse{
		ID:                record.ID,
		APIKeyID:          record.APIKeyID,
		GroupID:           record.GroupID,
		RequestID:         record.RequestID,
		UpstreamRequestID: record.UpstreamRequestID,
		Model:             record.Model,
		Prompt:            record.Prompt,
		Size:              record.Size,
		ResolutionTier:    record.ResolutionTier,
		AspectRatio:       record.AspectRatio,
		Quality:           record.Quality,
		OutputFormat:      record.OutputFormat,
		N:                 record.N,
		Status:            record.Status,
		StorageStatus:     record.StorageStatus,
		ImageCount:        record.ImageCount,
		Images:            publicImageGenerationImages(record.Images),
		FileSizeBytes:     record.FileSizeBytes,
		ErrorMessage:      record.ErrorMessage,
		CreatedAt:         record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt:       completed,
	}
}

func publicImageGenerationImages(images []service.ImageGenerationImage) []service.ImageGenerationImage {
	if len(images) == 0 {
		return nil
	}
	public := make([]service.ImageGenerationImage, len(images))
	for i := range images {
		public[i] = images[i]
		public[i].Path = ""
	}
	return public
}

func imageDownloadExt(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	default:
		return "png"
	}
}
