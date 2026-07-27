package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type WorkbenchHandler struct {
	service *service.WorkbenchService
	active  sync.Map
}

type activeWorkbenchGeneration struct {
	userID int64
	cancel context.CancelFunc
}

func NewWorkbenchHandler(service *service.WorkbenchService) *WorkbenchHandler {
	return &WorkbenchHandler{service: service}
}

func (h *WorkbenchHandler) Models(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	models, err := h.service.Models(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, models)
}

type addWorkbenchModelRequest struct {
	ModelID  string   `json:"model_id"`
	ModelIDs []string `json:"model_ids"`
	APIKeyID int64    `json:"api_key_id" binding:"required"`
}

func (h *WorkbenchHandler) AddModel(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	var req addWorkbenchModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	modelIDs := req.ModelIDs
	if strings.TrimSpace(req.ModelID) != "" {
		modelIDs = append(modelIDs, req.ModelID)
	}
	bindings, err := h.service.AddModels(c.Request.Context(), userID, modelIDs, req.APIKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, gin.H{"items": bindings})
}

func (h *WorkbenchHandler) HideModel(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	if err := h.service.HideModel(c.Request.Context(), userID, c.Param("id")); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"hidden": true})
}

type createWorkbenchConversationRequest struct {
	ModelBindingID  string `json:"model_binding_id"`
	ReasoningPreset string `json:"reasoning_preset"`
}

func (h *WorkbenchHandler) CreateConversation(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	var req createWorkbenchConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	conversation, err := h.service.CreateConversation(c.Request.Context(), userID, strings.TrimSpace(req.ModelBindingID), req.ReasoningPreset)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, conversation)
}

func (h *WorkbenchHandler) ListConversations(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	items, err := h.service.ListConversations(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *WorkbenchHandler) GetConversation(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	conversation, err := h.service.GetConversation(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, conversation)
}

type patchWorkbenchConversationRequest struct {
	Title           *string `json:"title"`
	ModelBindingID  *string `json:"model_binding_id"`
	ReasoningPreset *string `json:"reasoning_preset"`
}

func (h *WorkbenchHandler) UpdateConversation(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	var req patchWorkbenchConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	conversation, err := h.service.UpdateConversation(c.Request.Context(), userID, c.Param("id"), service.WorkbenchConversationPatch{
		Title: req.Title, ModelBindingID: req.ModelBindingID, ReasoningPreset: req.ReasoningPreset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, conversation)
}

func (h *WorkbenchHandler) DeleteConversation(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteConversation(c.Request.Context(), userID, c.Param("id")); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

type createWorkbenchMessageRequest struct {
	Content         string                        `json:"content"`
	ModelBindingID  string                        `json:"model_binding_id"`
	ReasoningPreset string                        `json:"reasoning_preset"`
	Attachments     []service.WorkbenchAttachment `json:"attachments"`
}

func (h *WorkbenchHandler) CreateMessage(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	var req createWorkbenchMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	turn, err := h.service.CreateTurn(
		c.Request.Context(), userID, c.Param("id"), req.Content,
		strings.TrimSpace(req.ModelBindingID), req.ReasoningPreset, req.Attachments,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, turn)
}

func (h *WorkbenchHandler) StreamGeneration(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	messageID := strings.TrimSpace(c.Param("id"))
	if messageID == "" {
		response.BadRequest(c, "Invalid generation ID")
		return
	}
	ctx, cancel := context.WithCancel(c.Request.Context())
	active := activeWorkbenchGeneration{userID: userID, cancel: cancel}
	if _, loaded := h.active.LoadOrStore(messageID, active); loaded {
		cancel()
		response.ErrorFrom(c, service.ErrWorkbenchGenerationStarted)
		return
	}
	defer func() {
		cancel()
		h.active.Delete(messageID)
	}()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher, _ := c.Writer.(http.Flusher)
	writeEvent := func(event string, payload any) error {
		data, _ := json.Marshal(payload)
		if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data); err != nil {
			return err
		}
		if flusher != nil {
			flusher.Flush()
		}
		return nil
	}
	if err := writeEvent("workbench.request.accepted", gin.H{"message_id": messageID}); err != nil {
		return
	}
	_ = writeEvent("workbench.task", gin.H{"id": "dispatch", "title": "正在调用模型", "status": "in_progress"})

	err := h.service.StreamGeneration(ctx, userID, messageID, func(event service.WorkbenchStreamEvent) error {
		return writeEvent(event.Type, event)
	})
	if err != nil {
		status := "failed"
		if errors.Is(err, context.Canceled) {
			status = "canceled"
		}
		_ = writeEvent("workbench.generation.failed", gin.H{"message_id": messageID, "status": status, "message": workbenchPublicError(err)})
		return
	}
	_ = writeEvent("workbench.task", gin.H{"id": "dispatch", "title": "回答已生成", "status": "completed"})
	_ = writeEvent("workbench.generation.completed", gin.H{"message_id": messageID})
}

func (h *WorkbenchHandler) CancelGeneration(c *gin.Context) {
	userID, ok := workbenchUserID(c)
	if !ok {
		return
	}
	messageID := strings.TrimSpace(c.Param("id"))
	if value, exists := h.active.Load(messageID); exists {
		if generation, ok := value.(activeWorkbenchGeneration); ok && generation.userID == userID {
			generation.cancel()
		}
	}
	if err := h.service.CancelGeneration(c.Request.Context(), userID, messageID); err != nil && !errors.Is(err, service.ErrWorkbenchMessageNotFound) {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"canceled": true})
}

func workbenchUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}
	return subject.UserID, true
}

func workbenchPublicError(err error) string {
	if err == nil {
		return ""
	}
	var gatewayErr *service.WorkbenchGatewayError
	if errors.As(err, &gatewayErr) && strings.TrimSpace(gatewayErr.Message) != "" {
		return gatewayErr.Message
	}
	if errors.Is(err, context.Canceled) {
		return "生成已停止"
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "生成失败，请稍后重试"
	}
	return message
}
